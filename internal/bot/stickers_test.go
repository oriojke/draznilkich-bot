package bot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLastWord(t *testing.T) {
	for _, tc := range []struct{ text, want string }{
		{"", ""},
		{" \t\n", ""},
		{"Ну ПРИВЕТ!  ", "привет"},
		{"первая\n«последняя»", "последня"},
		{"даааааа", "да"},
		{"Ну ДдДАаАа!!!", "да"},
		{"DAAAA", "da"},
		{"неееет", "нет"},
		{"мама", "мама"},
		{"111аа--бб", "111а--б"},
		{"кто-то", "кто-то"},
		{"привет пока", "пока"},
	} {
		if got := lastWord(tc.text); got != tc.want {
			t.Errorf("lastWord(%q) = %q, want %q", tc.text, got, tc.want)
		}
	}
}

func TestHandleMessageStickerReply(t *testing.T) {
	original := stickers
	stickers = map[string]string{"привет": "test-file-id"}
	t.Cleanup(func() { stickers = original })

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost || r.URL.Path != "/sendSticker" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var payload struct {
			ChatID  int64  `json:"chat_id"`
			Sticker string `json:"sticker"`
			Reply   struct {
				MessageID int64 `json:"message_id"`
			} `json:"reply_parameters"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if payload.ChatID != 42 || payload.Sticker != "test-file-id" || payload.Reply.MessageID != 123 {
			t.Errorf("unexpected sticker reply: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer server.Close()
	b := &Bot{baseURL: server.URL + "/", client: server.Client()}
	msg := &message{ID: 123, Text: "Ну ПРИВЕЕЕТ!"}
	msg.Chat.ID = 42
	if err := b.handleMessage(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("got %d requests, want 1", calls)
	}
	for _, text := range []string{"привет пока", "", " \t", "неизвестно"} {
		msg.Text = text
		if err := b.handleMessage(context.Background(), msg); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Errorf("unmatched messages caused extra requests: %d", calls)
	}
}
