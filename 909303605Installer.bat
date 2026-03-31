#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}    Zapret Installer v1.0.0${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

FOLDER_NAME="zapret[1.0.0]"

echo -e "${YELLOW}[1/4] Creating folder: $FOLDER_NAME${NC}"
mkdir -p "$FOLDER_NAME"

echo -e "${YELLOW}[2/4] Downloading zapret.main.exe...${NC}"
curl -L -o "$FOLDER_NAME/zapret.main.exe" "https://github.com/Likegenm/zapret/raw/refs/heads/main/zapret.main.exe"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ zapret.main.exe downloaded${NC}"
else
    echo -e "${RED}✗ Failed to download zapret.main.exe${NC}"
    exit 1
fi

echo -e "${YELLOW}[3/4] Downloading domains list...${NC}"
curl -L -o "$FOLDER_NAME/8908748636475#8475.txt" "https://raw.githubusercontent.com/Likegenm/zapret/refs/heads/main/8908748636475%238475.txt"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ domains list downloaded${NC}"
else
    echo -e "${RED}✗ Failed to download domains list${NC}"
    exit 1
fi

echo -e "${YELLOW}[4/4] Creating launcher script...${NC}"

cat > "$FOLDER_NAME/start.sh" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")"
echo "Starting Zapret Proxy..."
echo "Press Ctrl+C to stop"
./zapret.main.exe -domains "8908748636475#8475.txt" -v
EOF

cat > "$FOLDER_NAME/start.bat" << 'EOF'
@echo off
cd /d "%~dp0"
echo Starting Zapret Proxy...
echo Press Ctrl+C to stop
zapret.main.exe -domains "8908748636475#8475.txt" -v
pause
EOF

chmod +x "$FOLDER_NAME/zapret.main.exe" 2>/dev/null
chmod +x "$FOLDER_NAME/start.sh" 2>/dev/null

echo -e "${GREEN}✓ Launcher scripts created${NC}"
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}    Installation complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Folder: ${BLUE}$FOLDER_NAME${NC}"
echo ""
echo -e "${YELLOW}To start Zapret:${NC}"
echo -e "  Windows: ${BLUE}cd $FOLDER_NAME && start.bat${NC}"
echo -e "  Linux/Mac: ${BLUE}cd $FOLDER_NAME && ./start.sh${NC}"
echo ""
echo -e "${YELLOW}Or manually:${NC}"
echo -e "  ${BLUE}./zapret.main.exe -domains \"8908748636475#8475.txt\" -v${NC}"
echo ""
echo -e "${YELLOW}Then configure browser proxy:${NC}"
echo -e "  Proxy: ${BLUE}127.0.0.1:8080${NC}"
echo ""
