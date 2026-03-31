# 🚀 Zapret DPI Proxy

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-0078D4?style=for-the-badge&logo=windows)](https://github.com/Likegenm/zapret)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

<p align="center">
  <img src="https://skillicons.dev/icons?i=go" height="100">
</p>

<p align="center">
  <b>Профессиональный инструмент для обхода DPI (Deep Packet Inspection)</b><br>
  Легкий, быстрый и эффективный прокси-сервер на Go
</p>

---

## 📋 Оглавление
- [✨ Возможности](#-возможности)
- [⚡ Быстрый старт](#-быстрый-старт)
- [📖 Инструкция](#-инструкция)
- [🔧 Настройка браузера](#-настройка-браузера)
- [📁 Структура проекта](#-структура-проекта)
- [🛠️ Сборка из исходников](#️-сборка-из-исходников)
- [❓ Часто задаваемые вопросы](#-часто-задаваемые-вопросы)

---

## ✨ Возможности

| Функция | Описание |
|---------|----------|
| 🚀 **Высокая производительность** | Написан на Go, работает в разы быстрее аналогов |
| 🔒 **Обход DPI** | Фрагментация TLS-пакетов, обфускация SNI |
| 📋 **Список доменов** | Гибкая настройка через текстовый файл |
| 🌍 **Кроссплатформенность** | Работает на Windows, Linux, macOS |
| 💾 **Минимальный размер** | ~3-5 МБ в скомпилированном виде |
| 🖥️ **Простая установка** | Не требует установки Go для запуска |

---

## ⚡ Быстрый старт

### 📥 Установка за 1 минуту

```bash
# Скачайте установщик
git clone https://github.com/Likegenm/zapret.git
cd zapret

# Запустите установку (Linux/Mac)
chmod +x install.sh
./install.sh
```
# Или просто распакуйте архив и запустите start.bat (Windows)

## 📖 Инструкция
# 1️⃣ Запуск
Windows:

```cmd
zapret.main.exe -domains "8908748636475#8475.txt" -v
Linux / macOS:
```
```bash
./zapret.main.exe -domains "8908748636475#8475.txt" -v
```
# 2️⃣ Настрой браузер на прокси
Адрес: 127.0.0.1

Порт: 8080

# 3️⃣ Готово! 🎉
