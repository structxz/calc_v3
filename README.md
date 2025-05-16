# Распределенный вычислитель арифметических выражений[^1]

## Описание проекта

Этот проект представляет собой распределенную систему для вычисления математических выражений. Он состоит из:

- **Оркестратора** – управляет вычислениями и распределяет задачи между агентами.
- **Агентов** – выполняют вычисления по запросу оркестратора.
- **API** - предоставляет доступ через HTTP (REST) и gRPC для работы с вычислениями и задачами

## Функциональность

- Поддержка базовых арифметических операций (`+`, `-`, `*`, `/`).
- Возможность работы с выражениями, содержащими произвольное количество пробелов.
- Распределение вычислений между несколькими агентами.
- Логирование запросов и результатов вычислений.

## Структура проекта

```
project-root/
│
│── cmd/
│   ├── orchestrator/      # Главный модуль оркестратора
│   ├── agent/             # Главный модуль агента
│
│── configs/               # Файлы с конфигурациями
│
│── internal/
│   ├── app/               # Логика оркестратора
|   ├── auth/              # Логика хеширования пароля
│   ├── constants/         # Константы приложения
|   ├── db/
|       ├── sqlite         # Работа с базой данных SQLite
|   ├── jwtutil            # Логика работы JWT токенов
│   ├── logger/            # Логгер приложения
|   ├── middleware/        # Middleware с аутентификацией
|   ├── orchestrator       # Логика gRPC сервера
│   ├── worker/            # Логика вычислений
│
│── logs/              
│   │── agent/             # Логи агента
│   │── orchestrator/      # Логи оркестратора
│
│── tests/                 # Тесты проекта
│
│── pkg/                   # Логика вычислений
│
│── .env                   # Переменные окружения
│
│── docker-compose.yml     # Запуск приложения через Docker Compose
├── Dockerfile1            # Запуск оркестратора через Docker
├── Dockerfile2            # Запуск агента через Docker
├── Makefile               # Запуск Protoc для генерации .pb.go файлов
│── README.md              # Документация проекта
```

## Запуск проекта

### Локальный запуск

1. **Склонируйте репозиторий:**
   ```sh
   git clone https://github.com/structxz/calc_v3
   cd calc_v3
   ```
2. **Установите зависимости**
   ```sh
   go mod tidy
   ```
3. **Создайте файл `.env` и установите в нем переменные окружения**
   - Пример `.env` файла можете посмотреть в `.env.example`
   - Для того, чтобы создать свой `JWT_SECRET` запустите команду в терминале:

   ```sh
   openssl rand -base64 32
   ```
4. **Запустите систему:**
   ```sh
   go run ./cmd/orchestrator/main.go
   go run ./cmd/agent/main.go
   ```

### Использование Docker

#### Запуск

1. **Соберите контейнеры и запустите их (также необходимо установить зависимости и создать `.env` файл как и при локальном запуске):**

```sh
    docker compose up --build
```

2. **Остановка**

```sh
   docker compose stop
```

3. **Удаление**

```sh
docker compose down
```

## Взаимодействие с API

> Для того, чтобы взаимодействовать с выражениями, необходимо сначала зарегистрироваться и авторизоваться.

1. **Регистрация**

- `POST http://localhost:8080/api/v1/register`

  - Пример запроса

  ```json
  {
      "login": "testlogin",
      "password": "testpassword"
  }
  ```

2. **Авторизация**

- `POST http://localhost:8080/api/v1/login`

  - Пример запроса

  ```json
  {
      "login": "testlogin",
      "password": "testpassword"
  }
  ```

  - Пример ответа

  ```json
  {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY5NzQ4MDIsImlhdCI6MTc0Njk3NDY4Miwic3ViIjoiYWxleCJ9.o5ZWMgR2exVA1pOOE7jny0lhyUu2ti3X5t3h3TOp5GQ"
  }
  ```

3. **Отправка выражения на вычисление**

- `POST http://localhost:8080/api/v1/calculate`

  - Пример запроса

  ```json
  {
      "expression": "2+3"
  }
  ```

  - Пример ответа

  ```json
  {
      "id": "73ecc534-eb7b-4b12-83ec-4f441fbc98dc"
  }
  ```

4. **Получение информации о выражении по id**

- `GET http://localhost:8080/api/v1/expressions/73ecc534-eb7b-4b12-83ec-4f441fbc98dc`

  - Пример ответа

  ```json
  {
      "expression": {
            "id": "73ecc534-eb7b-4b12-83ec-4f441fbc98dc",
            "expression": "2+3",
            "status": "COMPLETE",
            "result": 5
      }
  }
  ```

5. **Получение всех выражений**

- `GET http://localhost:8080/api/v1/expressions`

  - Пример ответа (например вы отправили еще одно выражение)

  ```json
  {
      "expressions": [
        {
            "id": "415d1c9b-ef9d-45d6-b346-19505b9fc251",
            "expression": "3*3",
            "status": "COMPLETE",
            "result": 9
        },
        {
            "id": "73ecc534-eb7b-4b12-83ec-4f441fbc98dc",
            "expression": "2+3",
            "status": "COMPLETE",
            "result": 5,
        }
    ]
  }
  ```

### Взаимодействие через `curl`

**🔐 Регистрация пользователя**

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass"}'
```

**🔑 Вход в систему (получение JWT токена)**

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass"}'
```

💡 Примечание: В ответе будет JSON с токеном:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsIn..."
}
```

**➕ 3. Отправка выражения на вычисление**

```bash
curl -X POST http://localhost:8080/api/v1/calculate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"expression": "3 + 4"}'
```

`<TOKEN>` необходимо заменить на JWT из предыдущего шага.

**📋 4. Получение всех выражений пользователя**

```bash
curl -X GET http://localhost:8080/api/v1/expressions \
  -H "Authorization: Bearer <TOKEN>"
```

**🔎 5. Получение результата конкретного выражения**

```bash
curl -X GET http://localhost:8080/api/v1/expressions/1 \
  -H "Authorization: Bearer <TOKEN>"
```

`1` необходимо заменить на нужный ID выражения

---

## Схема работы системы

```text
+--------------------+                      +-------------------+                       
|    Client (User)   |                      |    Orchestrator   |
|    (REST / gRPC)   |                      +-------------------+
+---------+-----^----+                                |
     |          |                                     | (gRPC) 
     |  (REST)  |                                     | 
     |          |                           +---------v---------+  
+----v---------------+                      |      Agent 1      | 
|      REST API      |                      +-------------------+
|      (server)      |                                |
+---------------^----+                                |
     |          |                                     | (gRPC) 
     |  (REST)  |                                     |
     |          |                                     |      
+----v---------------+                      +---------v---------+ 
|  Database (SQLite) |                      |    Orchestrator   |
+--------------------+                      +-------------------+
```

1. Client (пользователь) отправляет запросы через REST API или gRPC.
2. Orchestrator управляет задачами и направляет их агентам для обработки.
3. Агенты выполняют вычисления и возвращают результаты.
4. Все задачи и результаты сохраняются в SQLite базе данных.

## Тестирование

1. **Запуск тестов:**
   ```sh
   go test ./test 
   ```

## Дальнейшие улучшения

- Оптимизация вычислений с кешированием.

[^1]: Проект при поддержке Яндекс Лицея и его лицеистов
