package bot

import (
	"context"
	"strings"
	"unicode"
)

// stickers maps lowercase words without repeated consecutive letters to Telegram sticker file_id values.
// Use file_id, not file_unique_id or a sticker pack URL.
var stickers = map[string]string{
	"да": "CAACAgIAAxkBAAMIasGD61eoGtC5MPOjlFjVgIQgu9AAAgVgAAKV6QhIoAlIGw-wyGY9BA",
	"da": "CAACAgIAAxkBAAMIasGD61eoGtC5MPOjlFjVgIQgu9AAAgVgAAKV6QhIoAlIGw-wyGY9BA",
	// "нет":  "CAACAgIAAxkBAAMXasGEfdvpZszHNUmykEPXucaXBDcAArIRAALMHilIrxZ-cC_u7Kk9BA",
	// "net":  "CAACAgIAAxkBAAMXasGEfdvpZszHNUmykEPXucaXBDcAArIRAALMHilIrxZ-cC_u7Kk9BA",
	// "nyet": "CAACAgIAAxkBAAMXasGEfdvpZszHNUmykEPXucaXBDcAArIRAALMHilIrxZ-cC_u7Kk9BA",
}

func lastWord(text string) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	word := strings.ToLower(strings.TrimFunc(words[len(words)-1], unicode.IsPunct))
	var normalized strings.Builder
	var previous rune
	for _, letter := range word {
		if letter != previous || !unicode.IsLetter(letter) {
			normalized.WriteRune(letter)
		}
		previous = letter
	}
	return normalized.String()
}

func (b *Bot) handleMessage(ctx context.Context, msg *message) error {
	if sticker := stickers[lastWord(msg.Text)]; sticker != "" {
		return b.call(ctx, "sendSticker", map[string]any{
			"chat_id":          msg.Chat.ID,
			"sticker":          sticker,
			"reply_parameters": map[string]any{"message_id": msg.ID},
		}, nil)
	}
	return nil
}
