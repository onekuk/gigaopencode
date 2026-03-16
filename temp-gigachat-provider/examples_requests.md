# Примеры запросов к OpenAI-compatible провайдеру

Все запросы выполняются к серверу на `http://localhost:8080/v1`. Authorization header обязателен для совместимости с OpenAI API, но сам токен может быть любым (например, `dummy_token`).

## Chat Completions

### Простой чат запрос

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "model": "GigaChat",
    "messages": [
      {"role": "user", "content": "Привет! Как дела?"}
    ],
    "max_tokens": 100,
    "temperature": 0.7
  }'
```

### Чат с историей сообщений

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "model": "GigaChat",
    "messages": [
      {"role": "system", "content": "Ты полезный помощник."},
      {"role": "user", "content": "Объясни квантовую физику простыми словами"},
      {"role": "assistant", "content": "Квантовая физика изучает поведение очень маленьких частиц..."},
      {"role": "user", "content": "А что такое квантовая запутанность?"}
    ],
    "max_tokens": 200,
    "temperature": 0.5
  }'
```

### Streaming чат

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "model": "GigaChat",
    "messages": [
      {"role": "user", "content": "Напиши короткую историю про кота"}
    ],
    "stream": true,
    "max_tokens": 300
  }'
```

## Text Completions

### Простое дополнение текста

```bash
curl -X POST http://localhost:8080/v1/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "model": "GigaChat",
    "prompt": "Однажды в студёную зимнюю пору",
    "max_tokens": 100,
    "temperature": 0.8
  }'
```

### Дополнение с параметрами

```bash
curl -X POST http://localhost:8080/v1/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "model": "GigaChat",
    "prompt": "Список преимуществ искусственного интеллекта:\n1.",
    "max_tokens": 150,
    "temperature": 0.3,
    "top_p": 0.9,
    "stop": ["\n\n"]
  }'
```

## Models

### Получить список доступных моделей

```bash
curl -X GET http://localhost:8080/v1/models \
  -H "Authorization: Bearer dummy_token"
```

### Получить информацию о конкретной модели

```bash
curl -X GET http://localhost:8080/v1/models/GigaChat \
  -H "Authorization: Bearer dummy_token"
```

## Embeddings

### Создать эмбеддинги для текста

```bash
curl -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "model": "text-embedding-ada-002",
    "input": "Это тестовый текст для создания эмбеддингов"
  }'
```

### Эмбеддинги для нескольких текстов

```bash
curl -X POST http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "model": "text-embedding-ada-002",
    "input": [
      "Первый текст",
      "Второй текст",
      "Третий текст"
    ]
  }'
```

## Image Generation

### Генерация изображения

```bash
curl -X POST http://localhost:8080/v1/images/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "prompt": "Красивый закат над морем в стиле импрессионизма",
    "n": 1,
    "size": "1024x1024"
  }'
```

### Генерация с дополнительными параметрами

```bash
curl -X POST http://localhost:8080/v1/images/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "prompt": "Футуристический город с летающими автомобилями",
    "model": "dall-e-3",
    "n": 2,
    "quality": "hd",
    "size": "1792x1024",
    "style": "vivid"
  }'
```

## Audio

### Транскрипция аудио

```bash
curl -X POST http://localhost:8080/v1/audio/transcriptions \
  -H "Content-Type: multipart/form-data" \
  -H "Authorization: Bearer dummy_token" \
  -F "file=@audio.mp3" \
  -F "model=whisper-1" \
  -F "language=ru"
```

### Перевод аудио

```bash
curl -X POST http://localhost:8080/v1/audio/translations \
  -H "Content-Type: multipart/form-data" \
  -H "Authorization: Bearer dummy_token" \
  -F "file=@audio_russian.mp3" \
  -F "model=whisper-1"
```

### Синтез речи

```bash
curl -X POST http://localhost:8080/v1/audio/speech \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "model": "tts-1",
    "input": "Привет! Это тест синтеза речи.",
    "voice": "alloy"
  }' \
  --output speech.mp3
```

## Moderation

### Проверка контента на нарушения

```bash
curl -X POST http://localhost:8080/v1/moderations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "input": "Это обычный безопасный текст для проверки.",
    "model": "text-moderation-latest"
  }'
```

## Files

### Загрузка файла

```bash
curl -X POST http://localhost:8080/v1/files \
  -H "Authorization: Bearer dummy_token" \
  -F "purpose=fine-tune" \
  -F "file=@dataset.jsonl"
```

### Список файлов

```bash
curl -X GET http://localhost:8080/v1/files \
  -H "Authorization: Bearer dummy_token"
```

### Получить информацию о файле

```bash
curl -X GET http://localhost:8080/v1/files/file-abc123 \
  -H "Authorization: Bearer dummy_token"
```

### Получить содержимое файла

```bash
curl -X GET http://localhost:8080/v1/files/file-abc123/content \
  -H "Authorization: Bearer dummy_token"
```

### Удалить файл

```bash
curl -X DELETE http://localhost:8080/v1/files/file-abc123 \
  -H "Authorization: Bearer dummy_token"
```

## Fine-tuning

### Создать задачу файн-тюнинга

```bash
curl -X POST http://localhost:8080/v1/fine_tuning/jobs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dummy_token" \
  -d '{
    "training_file": "file-abc123",
    "model": "gpt-3.5-turbo-0613",
    "hyperparameters": {
      "n_epochs": 3
    }
  }'
```

### Список задач файн-тюнинга

```bash
curl -X GET http://localhost:8080/v1/fine_tuning/jobs \
  -H "Authorization: Bearer dummy_token"
```

### Получить информацию о задаче

```bash
curl -X GET http://localhost:8080/v1/fine_tuning/jobs/ftjob-abc123 \
  -H "Authorization: Bearer dummy_token"
```

### Отменить задачу файн-тюнинга

```bash
curl -X POST http://localhost:8080/v1/fine_tuning/jobs/ftjob-abc123/cancel \
  -H "Authorization: Bearer dummy_token"
```

## Проверка статуса сервера

### Простая проверка доступности

```bash
curl -X GET http://localhost:8080/v1/models \
  -H "Authorization: Bearer test" \
  -w "\nHTTP Status: %{http_code}\nTime: %{time_total}s\n"
```

### Проверка с подробной информацией

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test" \
  -d '{"model":"GigaChat","messages":[{"role":"user","content":"тест"}],"max_tokens":10}' \
  -w "\nHTTP Status: %{http_code}\nResponse Time: %{time_total}s\nSize: %{size_download} bytes\n"
```

## Примечания

- Все запросы требуют заголовок `Authorization: Bearer <token>`
- Сервер автоматически управляет аутентификацией с GigaChat API
- Для multipart/form-data запросов (загрузка файлов) используйте соответствующий Content-Type
- Streaming запросы возвращают данные в формате Server-Sent Events