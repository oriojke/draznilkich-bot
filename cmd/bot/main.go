package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"tg-draznilka/internal/bot"
)

func main() {
	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if token == "" {
		slog.Error("задайте переменную окружения TELEGRAM_BOT_TOKEN")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := bot.New(token).Run(ctx); err != nil {
		slog.Error("бот остановлен с ошибкой", "error", err)
		os.Exit(1)
	}
	slog.Info("бот остановлен")
}
