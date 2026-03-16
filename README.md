# GigaChat Provider для OpenCode

Полная интеграция GigaChat (Сбер) с OpenCode через OpenAI-compatible API.

---

## История разработки

### Проблема

Нужно было подключить GigaChat (Сбер) к OpenCode так, чтобы:
- Не требовалась установка Go
- Всё работало из коробки
- Пользователь просто запускал OpenCode

### Исходные данные

- Go-сервер: https://gitverse.ru/kmpavloff/openai-provider-gigachat
- OpenCode поддерживает custom OpenAI-compatible провайдеры

---

## Что было сделано (пошагово)

### Шаг 1: Клонирование и изучение Go-сервера

```bash
git clone https://gitverse.ru/kmpavloff/openai-provider-gigachat.git temp-gigachat-provider
```

**Структура проекта:**
```
temp-gigachat-provider/
├── main.go              # Точка входа
├── router.go            # HTTP роутер с middleware
├── token_manager.go     # OAuth токены
├── gigachat_provider.go # Логика GigaChat API
└── handlers.go          # HTTP обработчики
```

**Ключевой вывод:** Сервер предоставляет OpenAI-compatible API на `localhost:8080/v1`

### Шаг 2: Компиляция бинарников

Скомпилирован сервер для всех платформ:

```bash
# Windows
go build -ldflags="-s -w" -o bin/gigachat-server-windows.exe .

# Linux x64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/gigachat-server-linux-x64 .

# Linux ARM64  
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/gigachat-server-linux-arm64 .

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/gigachat-server-darwin-x64 .

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/gigachat-server-darwin-arm64 .
```

**Результат:** Бинарники в `bin/` (6.5MB каждый)

### Шаг 3: Исправление ошибок SSL

**Проблема:** SSL-сертификат Сбера не распознавался системой

**Ошибка:**
```
tls: failed to verify certificate: x509: certificate signed by unknown authority
```

**Исправление в token_manager.go:**
```go
import "crypto/tls"

// Добавлено в NewTokenManager:
tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

httpClient: &http.Client{
    Transport: tr,
    Timeout:   5 * time.Minute,
}
```

**Исправление в gigachat_provider.go:**
```go
// Аналогичное изменение для HTTP клиента GigaChat
```

**Пересборка:** Все бинарники пересобраны с этими исправлениями.

### Шаг 4: Исправление авторизации

**Проблема:** OpenCode Desktop 1.2.24 не отправляет Authorization header

**Ошибка:**
```
Authorization header is required
```

**Исправление в router.go:**
```go
// Было:
if authHeader == "" {
    writeErrorResponse(..., "Authorization header is required")
    return
}

// Стало:
if authHeader != "" && !strings.HasPrefix(authHeader, "Bearer ") {
    writeErrorResponse(..., "Authorization header must start with 'Bearer '")
    return
}

if authHeader != "" {
    ctx := context.WithValue(r.Context(), contextKey("auth_token"), authHeader)
    r = r.WithContext(ctx)
}
```

**Результат:** Authorization header стал опциональным.

### Шаг 5: Создание конфигурации

**Создан файл `gigachat-config.json`:**
```json
{
  "authorization_key": "YOUR_BASE64_AUTH_KEY",
  "oauth_url": "https://ngw.devices.sberbank.ru:9443/api/v2/oauth",
  "scope": "GIGACHAT_API_PERS",
  "addr": "localhost",
  "port": "8080"
}
```

**Размещение:**
- Windows: `%APPDATA%\opencode\gigachat-config.json`
- Linux: `~/.config/opencode/gigachat-config.json`
- macOS: `~/Library/Application Support/opencode/gigachat-config.json`

**Создан файл `opencode.json` для проекта:**
```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "gigachat": {
      "name": "GigaChat",
      "npm": "@ai-sdk/openai-compatible",
      "options": {
        "baseURL": "http://localhost:8080/v1",
        "apiKey": "dummy"
      },
      "models": {
        "GigaChat-2-Max": {
          "name": "GigaChat Max 2",
          "tool_call": true,
          "limit": {
            "context": 8000,
            "output": 4096
          }
        }
      }
    }
  },
  "model": "gigachat/GigaChat-2-Max"
}
```

### Шаг 6: Тестирование

**Проверка сервера:**
```bash
curl http://localhost:8080/v1/models -H "Authorization: Bearer test"
# Результат: список из 14 моделей
```

**Проверка чата:**
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"GigaChat-2-Max","messages":[{"role":"user","content":"Привет"}]}'
# Результат: ответ от GigaChat
```

### Шаг 7: Отладка проблемы с контекстом

**Проблема:** Бесконечные повторные запросы от OpenCode

**Анализ логов:**
```
Content-Length: 599152  ← 600KB данных!
context canceled         ← OpenCode закрывает соединение
Status: 500              ← Ошибка сервера
```

**Причина:**
1. OpenCode отправляет огромный контекст (история сообщений)
2. У OpenCode короткий таймаут (~700ms)
3. GigaChat API не успевает обработать
4. OpenCode отменяет запрос
5. Цикл повторяется

**Решение:** Уменьшение лимита контекста в opencode.json
```json
{
  "models": {
    "GigaChat-2-Max": {
      "limit": {
        "context": 8000,   // Было 64000
        "output": 4096     // Было 8192
      }
    }
  }
}
```

### Шаг 8: Создание скриптов запуска

Созданы удобные bat-файлы:

**`start-gigachat-simple.bat`:**
```batch
@echo off
cd /d "C:\Users\%USERNAME%\Desktop\opencodegiga"
.\bin\gigachat-server-windows.exe
pause
```

**`launch-opencode-with-gigachat.bat`:**
```batch
@echo off
start "GigaChat Server" cmd /k "...\gigachat-server-windows.exe"
timeout /t 5
cd /d "...\test-opencode-project"
start "" opencode
```

---

## Итоговая архитектура

```
┌──────────────┐     HTTP      ┌─────────────────────┐     HTTPS      ┌─────────────┐
│   OpenCode   │ ────────────> │ gigachat-server     │ ────────────> │  GigaChat   │
│  (клиент)    │  OpenAI API   │ (localhost:8080)    │  GigaChat API │   (Сбер)    │
└──────────────┘               └─────────────────────┘               └─────────────┘
                                       │
                                       │
                                       ▼
                              config.json
                              - authorization_key
                              - oauth_url
                              - scope
```

---

## Ключевые изменения кода

### 1. router.go
```go
// Сделан Authorization опциональным
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader != "" && !strings.HasPrefix(authHeader, "Bearer ") {
            writeErrorResponse(...)
            return
        }
        if authHeader != "" {
            ctx := context.WithValue(r.Context(), contextKey("auth_token"), authHeader)
            r = r.WithContext(ctx)
        }
        next(w, r)
    }
}
```

### 2. token_manager.go
```go
// Добавлено игнорирование SSL
func NewTokenManager(config *Config, logger *Logger) *TokenManager {
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    return &TokenManager{
        httpClient: &http.Client{
            Transport: tr,
            Timeout:   5 * time.Minute,
        },
    }
}
```

### 3. gigachat_provider.go
```go
// Аналогичное изменение для HTTP клиента
```

---

## Файлы проекта

```
opencodegiga/
├── bin/                              # Бинарники (6.5MB каждый)
│   ├── gigachat-server-windows.exe
│   ├── gigachat-server-linux-x64
│   ├── gigachat-server-linux-arm64
│   ├── gigachat-server-darwin-x64
│   └── gigachat-server-darwin-arm64
│
├── temp-gigachat-provider/           # Исходники Go (модифицированные)
│   ├── main.go
│   ├── router.go                     # ← Изменён
│   ├── token_manager.go              # ← Изменён
│   ├── gigachat_provider.go          # ← Изменён
│   └── handlers.go
│
├── src/                              # TypeScript (опционально)
│   └── index.ts
│
├── config.json                       # Конфиг сервера
├── logs/                             # Логи работы
│
├── package.json                      # NPM
├── tsconfig.json
│
├── README.md                         # Этот файл
├── INSTALL.md                        # Инструкция по установке
├── DOCUMENTATION.md                  # Документация
├── QUICKSTART.md                     # Быстрый старт
│
└── *.bat                             # Скрипты запуска
    ├── start-gigachat-simple.bat
    ├── launch-opencode-with-gigachat.bat
    ├── test-gigachat.bat
    └── debug-opencode.bat
```

---

## Установка (быстрый старт)

### 1. Получить API ключ

1. Зарегистрироваться на https://developers.sber.ru
2. Создать проект
3. Скопировать Authorization Key

### 2. Создать конфиг

**Файл:** `C:\Users\%USERNAME%\AppData\Roaming\opencode\gigachat-config.json`

```json
{
  "authorization_key": "ВАШ_КЛЮЧ",
  "oauth_url": "https://ngw.devices.sberbank.ru:9443/api/v2/oauth",
  "scope": "GIGACHAT_API_PERS"
}
```

### 3. Создать конфиг OpenCode

**Файл:** `opencode.json` в папке проекта

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "gigachat": {
      "name": "GigaChat",
      "npm": "@ai-sdk/openai-compatible",
      "options": {
        "baseURL": "http://localhost:8080/v1",
        "apiKey": "dummy"
      },
      "models": {
        "GigaChat-2-Max": {
          "name": "GigaChat Max 2",
          "tool_call": true,
          "limit": {
            "context": 8000,
            "output": 4096
          }
        }
      }
    }
  },
  "model": "gigachat/GigaChat-2-Max"
}
```

### 4. Запустить

**Окно 1 - Сервер:**
```bash
.\bin\gigachat-server-windows.exe
```

**Окно 2 - OpenCode:**
```bash
opencode
```

---

## Исправленные проблемы

| Проблема | Причина | Решение |
|----------|---------|---------|
| SSL ошибка | Сертификат Сбера | `InsecureSkipVerify: true` |
| Authorization required | OpenCode не шлёт header | Сделать опциональным |
| Бесконечные запросы | Большой контекст + таймаут | Уменьшить limit context |
| Context canceled | OpenCode закрывает соединение | Оптимизация таймаутов |

---

## Время разработки

- Исследование: ~30 мин
- Компиляция: ~15 мин  
- SSL fix: ~20 мин
- Auth fix: ~25 мин
- Тестирование: ~45 мин
- Документация: ~30 мин

**Итого:** ~2.5 часа

---

## Статус

✅ **Работает:**
- Сервер запускается локально
- OpenCode подключается
- Ответы приходят в чат
- Все модели GigaChat доступны

⚠️ **Ограничения:**
- Нужно запускать сервер отдельно
- Контекст ограничен 8000 токенами
- SSL проверка отключена (тест)

---

## Лицензия

MIT

## Благодарности

- [openai-provider-gigachat](https://gitverse.ru/kmpavloff/openai-provider-gigachat) - оригинальный сервер
- [OpenCode](https://opencode.ai) - AI-powered development tool
