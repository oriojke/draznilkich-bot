package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Bot struct {
	baseURL string
	client  *http.Client
}

func New(token string) *Bot {
	return &Bot{
		baseURL: "https://api.telegram.org/bot" + token + "/",
		client:  &http.Client{Timeout: 40 * time.Second},
	}
}

type update struct {
	ID      int64    `json:"update_id"`
	Message *message `json:"message"`
}

type message struct {
	ID   int64  `json:"message_id"`
	Text string `json:"text"`
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
}

// Run receives messages via long polling until the context is cancelled.
func (b *Bot) Run(ctx context.Context) error {
	if err := b.call(ctx, "getMe", struct{}{}, nil); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("проверка токена: %w", err)
	}
	slog.Info("бот запущен")
	var offset int64
	for ctx.Err() == nil {
		var updates []update
		err := b.call(ctx, "getUpdates", map[string]any{
			"offset": offset, "timeout": 30, "allowed_updates": []string{"message"},
		}, &updates)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Error("не удалось получить сообщения", "error", err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(5 * time.Second):
				continue
			}
		}
		for _, u := range updates {
			if ctx.Err() != nil {
				return nil
			}
			if u.Message != nil && u.Message.Text != "" {
				if err := b.handleMessage(ctx, u.Message); err != nil && ctx.Err() == nil {
					slog.Error("не удалось отправить ответ", "error", err)
				}
			}
			// Failed replies are logged and skipped; this scaffold has no retry queue.
			offset = u.ID + 1
		}
	}
	return nil
}

func (b *Bot) call(ctx context.Context, method string, payload any, result any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.baseURL+method, bytes.NewReader(data))
	if err != nil {
		return errors.New("не удалось создать запрос к Telegram")
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := b.client.Do(req)
	if err != nil {
		// net/http errors include the URL, which contains the bot token.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return fmt.Errorf("Telegram %s: %w", method, urlErr.Err)
		}
		return errors.New("ошибка соединения с Telegram")
	}
	defer res.Body.Close()
	var envelope struct {
		OK          bool            `json:"ok"`
		Description string          `json:"description"`
		Result      json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(res.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("Telegram %s (HTTP %d): некорректный ответ", method, res.StatusCode)
	}
	if !envelope.OK || res.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram %s (HTTP %d): %s", method, res.StatusCode, envelope.Description)
	}
	if result != nil {
		return json.Unmarshal(envelope.Result, result)
	}
	return nil
}
