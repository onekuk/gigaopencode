# OpenAI-compatible Provider для GigaChat API

OpenAI-совместимый HTTP сервер, проксирующий запросы к GigaChat API Сбербанка.

## Функции

- Полная совместимость с OpenAI API
- Автоматическое управление access tokens с обновлением
- Поддержка всех основных endpoints:
  - Chat Completions (включая streaming)
  - Text Completions
  - Embeddings
  - Models
  - Files
  - Fine-tuning
  - Images
  - Audio
  - Moderation
- CORS поддержка
- Конфигурация через JSON файл

## Установка

1. Склонируйте репозиторий
2. Скопируйте `config.example.json` в `config.json`
3. Добавьте ваш authorization key в `config.json`

```bash
cp config.example.json config.json
# Отредактируйте config.json с вашим ключом
```

## Конфигурация

Создайте файл `config.json`:

```json
{
  "authorization_key": "ваш_base64_authorization_key",
  "oauth_url": "https://ngw.devices.sberbank.ru:9443/api/v2/oauth",
  "scope": "GIGACHAT_API_PERS",
  "addr": "localhost",
  "port": "8080"
}
```

### Параметры конфигурации

- `authorization_key` - ключ авторизации GigaChat API (обязательный)
- `oauth_url` - URL для получения токена доступа (по умолчанию: `https://ngw.devices.sberbank.ru:9443/api/v2/oauth`)
- `scope` - область доступа API (по умолчанию: `GIGACHAT_API_PERS`)
- `addr` - адрес сервера (по умолчанию: `localhost`)
- `port` - порт сервера (по умолчанию: `8080`)

## Запуск

```bash
go run .
```

Или с кастомными параметрами:

```bash
ADDR=0.0.0.0 PORT=9000 CONFIG_PATH=config.json LOG_LEVEL=DEBUG go run .
```

### Переменные окружения

- `ADDR` - адрес сервера (переопределяет значение из config.json)
- `PORT` - порт сервера (переопределяет значение из config.json)
- `CONFIG_PATH` - путь к файлу конфигурации (по умолчанию: config.json)
- `LOG_LEVEL` - уровень логирования: DEBUG, INFO, WARN, ERROR (по умолчанию: INFO)

Примечание: переменные окружения `ADDR` и `PORT` имеют приоритет над значениями в конфигурационном файле.

## Использование

Сервер запускается на `http://localhost:8080/v1` и полностью совместим с OpenAI API.

Пример запроса:

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "model": "GigaChat",
    "messages": [{"role": "user", "content": "Привет!"}]
  }'
```

Примечание: Authorization header обязателен для совместимости с OpenAI клиентами, но сам токен не используется - аутентификация происходит через конфигурацию сервера.

## Логирование

Сервер поддерживает структурированное логирование с четырьмя уровнями:

- **DEBUG** - подробная информация о всех запросах/ответах, токенах, телах запросов
- **INFO** - основные события: запуск сервера, завершение запросов, обновление токенов
- **WARN** - предупреждения о потенциальных проблемах
- **ERROR** - ошибки выполнения

Примеры логов:

```
[INFO ] 2024-01-15 10:30:45 Starting OpenAI-compatible provider for GigaChat
[DEBUG] 2024-01-15 10:30:46 Request: POST https://gigachat.devices.sberbank.ru/api/v1/chat/completions, Token: eyJjdH...W9hQ
[INFO ] 2024-01-15 10:30:47 Request completed: POST /chat/completions, Status: 200, Duration: 1.2s
```

Управление уровнем логирования:

```bash
LOG_LEVEL=DEBUG go run .    # Подробные логи
LOG_LEVEL=INFO go run .     # Стандартные логи (по умолчанию)
LOG_LEVEL=ERROR go run .    # Только ошибки
```

## Архитектура

Проект построен по модульному принципу с разделением ответственности:

- `Config` - управление конфигурацией сервера (файл + переменные окружения)
- `Logger` - единая система структурированного логирования
- `TokenManager` - автоматическое управление получением и обновлением access_token
- `GigaChatProvider` - проксирование запросов к GigaChat API с конвертацией форматов
- `Converters` - преобразование моделей данных между форматами OpenAI и GigaChat
- `HTTPHandlers` - обработка HTTP запросов для всех endpoints
- `Router` - маршрутизация запросов и middleware (CORS, auth, logging)

Подробная диаграмма архитектуры доступна в файле [arch.puml](arch.puml) (C4 модель).


## Подключение провайдера в opencode

[OpenCode](https://opencode.ai/docs/) это CLI Agent, который может использовать разные LLM модели совместимые с OpenAI-like API. Чтобы подключить новый провайдер нужно выполнить следующие шаги:

1. добавить новый provider


```bash
opencode auth login
```

вводим имя провайдера `"salutedevices` который мы будем использовать далее в файле `opencode.json`

2. создать файл `opencode.json` в корне проекта в котором будет запускаться opencode

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "salutedevices": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "GigaChat AI",
      "options": {
        "baseURL": "http://localhost:8080/v1",
        "apiKey": "test_api_key"
      },
      "models": {
        "GigaChat-2-Max": {
          "name": "GigaChat Max 2"
        }
      }
    }
  }
}
```

apiKey можно указать любой, так как сервером будет получаться реальный `access_token` для доступа к GigaChat. Ссылка на [инструкцию по подключению CustomProvider в opencode](https://opencode.ai/docs/providers/#custom-provider)


## Подключение провайдера в crush

[Crush](https://github.com/charmbracelet/crush) аналогичный opencode cli Agent. Ниже представлен конфигурационный файл по добавлению провайдера. 

1. Добавить файл конфигурации в `$HOME/.config/crush/crush.json` следующее описание: 

```json
{
  "providers": {
    "gigachat": {
      "type": "openai",
      "base_url": "http://localhost:8080/v1",
      "api_key": "dummy",
      "models": [
         {
           "id": "GigaChat-2-Max",
           "name": "GigaChat Max 2",
           "context_window": 64000,
           "default_max_tokens": 5000
         },
         {
           "id": "GigaChat-2",
           "name": "GigaChat 2",
           "context_window": 64000,
           "default_max_tokens": 5000
         },
         {
           "id": "GigaChat-2-Pro",
           "name": "GigaChat Pro 2",
           "context_window": 64000,
           "default_max_tokens": 5000
         }
      ]
    }
  }
}
```

2. Запустите `openai-provider-gigachat` и `crush`. В crush смените используемую модель через меню `Ctrl+P`.

Подробное описание работы с crush можно найти на [github'е](https://github.com/charmbracelet/crush).