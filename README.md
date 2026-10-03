# tg-draznilka

Заготовка Telegram-бота на Go 1.22+ без внешних зависимостей.
Получает сообщения через long polling, отвечает стикером по последнему слову,
завершается по Ctrl+C. HTTP-клиент использует [Telegram Bot API](https://core.telegram.org/bots/api).

## Запуск

Создайте бота через [@BotFather](https://t.me/BotFather) командой `/newbot` и получите токен.

PowerShell:

```powershell
$env:TELEGRAM_BOT_TOKEN = "ваш_токен"
go run ./cmd/bot
```

Linux / macOS:

```sh
export TELEGRAM_BOT_TOKEN="ваш_токен"
go run ./cmd/bot
```

`.env.example` описывает нужную переменную. При запуске через `go run` файлы `.env` автоматически не загружаются:
передавайте переменную через оболочку или настройки запуска IDE.
После запуска откройте чат с ботом и отправьте сообщение со словом из мапы в конце.
Запускайте один экземпляр бота с одним токеном. Если ранее был настроен webhook,
удалите его через `deleteWebhook` перед использованием long polling.

## Docker

Нужны Docker с поддержкой Linux-контейнеров и Docker Compose v2.

1. Скопируйте `.env.example` в `.env` (`Copy-Item .env.example .env` в PowerShell
   или `cp .env.example .env` в Linux / macOS).
2. Замените значение `TELEGRAM_BOT_TOKEN` в `.env` на токен своего бота.
3. Запустите:

```sh
docker compose up -d --build
docker compose logs -f bot
```

Compose автоматически читает `.env`. Переменная из оболочки имеет приоритет
над значением в `.env`. Не запускайте одновременно локального бота с тем же токеном.

Остановка:

```sh
docker compose down
```

После изменения мапы стикеров или кода повторите `docker compose up -d --build`.
Порты и тома не нужны: бот использует исходящие HTTPS-запросы и не хранит данные.
Образ собирается в два этапа, содержит статический бинарник и CA-сертификаты,
запускается от непривилегированного пользователя. Токен передаётся только при запуске.

Без Compose (сначала задайте переменную окружения в оболочке):

```sh
docker build -t tg-draznilka .
docker run --rm --env TELEGRAM_BOT_TOKEN tg-draznilka
```

## Структура

```text
cmd/bot/main.go       — настройки, запуск и завершение
internal/bot/bot.go   — Telegram API, получение сообщений и обработчики
internal/bot/stickers.go — мапа слов и отправка стикеров
.env.example         — пример настройки окружения
Dockerfile           — сборка образа
compose.yaml         — запуск контейнера
```

Заполните мапу `stickers` в `internal/bot/stickers.go`: ключ — слово в нижнем
регистре, значение — Telegram `file_id` стикера (не `file_unique_id` и не ссылка
на набор).

Сообщение делится по пробельным символам. У последнего слова убираются знаки
препинания по краям и приводится к нижнему регистру: `Ну ПРИВЕТ!` ищет ключ
`привет`. При совпадении бот отправляет стикер ответом на исходное сообщение
через [sendSticker](https://core.telegram.org/bots/api#sendsticker).
Без совпадения бот молчит.
Для обработки обычных сообщений в группе отключите privacy mode у бота через
@BotFather (`/setprivacy`), затем добавьте бота в группу заново.

Очереди повторной отправки и постоянного хранилища пока нет:
ошибка отправки записывается в лог, после чего бот переходит к следующему сообщению.

## Проверка и сборка

```sh
go test ./...
go vet ./...
go build -o bin/bot ./cmd/bot
```

В Windows для сборки используйте `go build -o bin/bot.exe ./cmd/bot`.

## Публикация на GitHub

Создайте пустой репозиторий на GitHub без автоматически добавленных файлов.
В корне проекта выполните, заменив `YOUR_USERNAME` на свой логин:

```sh
git init -b main
git add .
git diff --cached --stat
git commit -m "Initial Telegram bot with Docker support"
git remote add origin https://github.com/YOUR_USERNAME/tg-draznilka.git
git push -u origin main
```

`.env`, бинарники и настройки IDE исключены через `.gitignore`.
В Docker-контекст попадают только `go.mod` и исходники из `cmd` и `internal`.
Не записывайте токен в исходники. Идентификаторы стикеров в мапе будут опубликованы
вместе с кодом; для другого бота следует получить его собственные `file_id`.

