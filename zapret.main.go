package main

import (
    "bufio"
    "flag"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "strings"
    "sync"
    "sync/atomic"
    "syscall"
    "time"
)

type DPIProxy struct {
    listenAddr    string
    fragmentSize  int
    fragmentDelay time.Duration
    targetPorts   map[int]bool
    targetDomains map[string]bool
    verbose       bool
    stats         *Stats
    connections   sync.Map
    wg            sync.WaitGroup
    shutdown      atomic.Bool
}

type Stats struct {
    totalConns      int64
    activeConns     int64
    bytesTransferred int64
    fragmentedConns int64
}

type Connection struct {
    id        uint64
    client    net.Conn
    target    net.Conn
    startTime time.Time
    bytesSent int64
    bytesRecv int64
    fragmented bool
    remoteAddr string
}

func NewDPIProxy() *DPIProxy {
    return &DPIProxy{
        targetPorts:   make(map[int]bool),
        targetDomains: make(map[string]bool),
        stats:         &Stats{},
        fragmentSize:  5,
        fragmentDelay: 10 * time.Millisecond,
    }
}

func (p *DPIProxy) LoadDomainsFromFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("failed to open %s: %v", filename, err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    count := 0

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        line = strings.Trim(line, "\r\n\t ")
        p.targetDomains[line] = true
        count++
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("read error: %v", err)
    }

    log.Printf("[+] Loaded %d domains from %s", count, filename)
    return nil
}

func (p *DPIProxy) LoadPortsFromFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    count := 0

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        var port int
        fmt.Sscanf(line, "%d", &port)
        if port > 0 && port < 65536 {
            p.targetPorts[port] = true
            count++
        }
    }

    log.Printf("[+] Loaded %d ports from file", count)
    return nil
}

func (p *DPIProxy) Run() error {
    listener, err := net.Listen("tcp", p.listenAddr)
    if err != nil {
        return fmt.Errorf("failed to listen: %v", err)
    }
    defer listener.Close()

    log.Printf("========================================")
    log.Printf("DPI Proxy Professional - Domain Edition")
    log.Printf("========================================")
    log.Printf("[+] Server: %s", p.listenAddr)
    log.Printf("[+] Fragment size: %d bytes", p.fragmentSize)
    log.Printf("[+] Target domains: %d", len(p.targetDomains))
    log.Printf("[+] Target ports: %d", len(p.targetPorts))
    log.Printf("")

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("\n[-] Shutting down...")
        p.shutdown.Store(true)
        listener.Close()
    }()

    go p.statsReporter()

    var connID uint64
    for {
        client, err := listener.Accept()
        if err != nil {
            if p.shutdown.Load() {
                break
            }
            continue
        }

        atomic.AddInt64(&p.stats.totalConns, 1)
        atomic.AddInt64(&p.stats.activeConns, 1)

        connID++
        p.wg.Add(1)
        go p.handleClient(client, connID)
    }

    p.wg.Wait()
    log.Printf("[+] Shutdown complete")
    return nil
}

func (p *DPIProxy) handleClient(client net.Conn, id uint64) {
    defer p.wg.Done()
    defer client.Close()
    defer atomic.AddInt64(&p.stats.activeConns, -1)

    client.SetDeadline(time.Now().Add(10 * time.Second))

    buf := make([]byte, 4096)
    n, err := client.Read(buf)
    if err != nil {
        return
    }

    client.SetDeadline(time.Time{})
    request := string(buf[:n])

    if strings.HasPrefix(request, "CONNECT") {
        p.handleHTTPS(client, request, id)
    } else if strings.HasPrefix(request, "GET") || strings.HasPrefix(request, "POST") ||
        strings.HasPrefix(request, "PUT") || strings.HasPrefix(request, "DELETE") ||
        strings.HasPrefix(request, "HEAD") || strings.HasPrefix(request, "OPTIONS") {
        p.handleHTTP(client, buf[:n], request, id)
    }
}

func (p *DPIProxy) handleHTTP(client net.Conn, data []byte, request string, id uint64) {
    var host string
    lines := strings.Split(request, "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "Host:") {
            host = strings.TrimSpace(strings.TrimPrefix(line, "Host:"))
            break
        }
    }

    if host == "" {
        return
    }

    if strings.Contains(host, ":") {
        host = strings.Split(host, ":")[0]
    }

    targetAddr := fmt.Sprintf("%s:80", host)
    target, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
    if err != nil {
        return
    }
    defer target.Close()

    target.Write(data)

    conn := &Connection{
        id:         id,
        client:     client,
        target:     target,
        startTime:  time.Now(),
        remoteAddr: host,
    }

    go p.pipe(client, target, conn, true)
    p.pipe(target, client, conn, false)
}

func (p *DPIProxy) handleHTTPS(client net.Conn, request string, id uint64) {
    hostPort := strings.Split(request, " ")[1]
    host, port, err := net.SplitHostPort(hostPort)
    if err != nil {
        host = hostPort
        port = "443"
    }

    targetPort := 443
    if p, err := net.LookupPort("tcp", port); err == nil {
        targetPort = p
    }

    client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

    targetAddr := fmt.Sprintf("%s:%d", host, targetPort)
    target, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
    if err != nil {
        return
    }
    defer target.Close()

    conn := &Connection{
        id:         id,
        client:     client,
        target:     target,
        startTime:  time.Now(),
        remoteAddr: host,
    }

    needFragment := false

    if p.targetPorts[targetPort] {
        needFragment = true
    }

    for domain := range p.targetDomains {
        if strings.Contains(host, domain) {
            needFragment = true
            break
        }
    }

    if needFragment {
        atomic.AddInt64(&p.stats.fragmentedConns, 1)
        conn.fragmented = true
        p.handleFragmentedTLS(conn)
    } else {
        go p.pipe(client, target, conn, true)
        p.pipe(target, client, conn, false)
    }
}

func (p *DPIProxy) handleFragmentedTLS(conn *Connection) {
    buf := make([]byte, 65535)
    n, err := conn.target.Read(buf)
    if err != nil {
        return
    }

    data := buf[:n]

    if len(data) < 5 || data[0] != 0x16 {
        conn.client.Write(data)
        go p.pipe(conn.target, conn.client, conn, true)
        p.pipe(conn.client, conn.target, conn, false)
        return
    }

    if len(data) <= p.fragmentSize {
        conn.client.Write(data)
    } else {
        conn.client.Write(data[:p.fragmentSize])
        time.Sleep(p.fragmentDelay)
        conn.client.Write(data[p.fragmentSize:])
    }

    go p.pipe(conn.target, conn.client, conn, true)
    p.pipe(conn.client, conn.target, conn, false)
}

func (p *DPIProxy) pipe(dst, src net.Conn, conn *Connection, isClient bool) {
    buf := make([]byte, 32768)

    for {
        n, err := src.Read(buf)
        if err != nil {
            break
        }

        _, err = dst.Write(buf[:n])
        if err != nil {
            break
        }

        atomic.AddInt64(&p.stats.bytesTransferred, int64(n))
        if isClient {
            conn.bytesSent += int64(n)
        } else {
            conn.bytesRecv += int64(n)
        }
    }
}

func (p *DPIProxy) statsReporter() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        if p.shutdown.Load() {
            return
        }

        total := atomic.LoadInt64(&p.stats.totalConns)
        active := atomic.LoadInt64(&p.stats.activeConns)
        bytesXfer := atomic.LoadInt64(&p.stats.bytesTransferred)
        fragmented := atomic.LoadInt64(&p.stats.fragmentedConns)

        log.Printf("[Stats] Total: %d, Active: %d, Xfer: %.2f MB, Frag: %d",
            total, active, float64(bytesXfer)/1024/1024, fragmented)
    }
}

func main() {
    listenAddr := flag.String("listen", "127.0.0.1:8080", "Listen address")
    fragmentSize := flag.Int("fragment", 5, "Fragment size (bytes)")
    fragmentDelayMs := flag.Int("delay", 10, "Fragment delay (ms)")
    portsFile := flag.String("ports", "", "Ports file (optional)")
    domainsFile := flag.String("domains", "", "Domains file (required)")
    verbose := flag.Bool("v", false, "Verbose output")

    flag.Parse()

    if *domainsFile == "" {
        log.Fatal("ERROR: -domains file is required")
        return
    }

    proxy := NewDPIProxy()
    proxy.listenAddr = *listenAddr
    proxy.fragmentSize = *fragmentSize
    proxy.fragmentDelay = time.Duration(*fragmentDelayMs) * time.Millisecond
    proxy.verbose = *verbose

    if err := proxy.LoadDomainsFromFile(*domainsFile); err != nil {
        log.Fatalf("Failed to load domains: %v", err)
    }

    if *portsFile != "" {
        if err := proxy.LoadPortsFromFile(*portsFile); err != nil {
            log.Printf("Warning: failed to load ports: %v", err)
        }
    }

    if len(proxy.targetDomains) == 0 {
        log.Fatal("ERROR: domains file is empty")
    }

    if err := proxy.Run(); err != nil {
        log.Fatal(err)
    }
}
