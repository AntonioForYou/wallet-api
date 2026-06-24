## Запуск приложения
Для запуска используется Docker Compose:
```bash
docker compose up
```
## Переменные окружения config.env

| Переменная | Описание | Значение (для теста) |
| :--- | :--- |:---------------------|
| `POSTGRES_USER` | Пользователь БД (для Docker) | `wallet_user`        |
| `POSTGRES_PASSWORD` | Пароль БД (для Docker) | `wallet_password`    |
| `POSTGRES_DB` | Имя БД (для Docker) | `wallet_db`          |
| `DB_USER` | Пользователь БД для приложения | `wallet_user`        |
| `DB_PASSWORD` | Пароль БД для приложения | `wallet_password`    |
| `DB_NAME` | Имя базы данных | `wallet_db`          |
| `DB_HOST` | Хост БД | `postgres`           |
| `DB_PORT` | Порт БД | `5432`               |
| `DB_SSLMODE` | Режим SSL соединения | `disable`            |
| `SERVER_PORT` | Порт API приложения | `8080`               |
| `WORKER_POOL_SIZE` | Количество воркеров | `100`                |
| `WORKER_BUFFER_SIZE` | Размер буфера очереди задач | `500`                |