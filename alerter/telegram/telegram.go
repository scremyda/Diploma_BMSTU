package telegram

import (
	"context"
	"diploma/alerter/repo"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	BotToken string `yaml:"bot_token"`
	ChatID   int64  `yaml:"chat_id"`
}

type TelegramBot struct {
	repo repo.QueueRepo
	bot  *tgbotapi.BotAPI
	conf Config
}

func (tb *TelegramBot) ProcessQueue(ctx context.Context) {
	ticker := time.NewTicker(tb.repo.PollInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stop events processing")
			return
		case <-ticker.C:
			batchID, err := tb.repo.NextBatch(ctx)
			if err != nil {
				log.Printf("Error receiving new batch: %v", err)
				continue
			}
			if batchID <= 0 {
				// Если событий нет, просто ждем следующего тикера.
				continue
			}

			events, err := tb.repo.GetBatchEvents(ctx, batchID)
			if err != nil {
				log.Printf("Error getting batch events for batch %d: %v", batchID, err)
				tb.repo.FinishBatch(ctx, batchID)
				continue
			}

			for _, event := range events {
				text := fmt.Sprintf("Error for event %s: %s", event.Target, event.Message)
				if err := tb.SendMessage(text); err != nil {
					log.Printf("Error sending message to Telegram: %v", err)
				} else {
					log.Printf("Message sent: %s", text)
				}
			}

			if err := tb.repo.FinishBatch(ctx, batchID); err != nil {
				log.Printf("Error finishing batch %d: %v", batchID, err)
			} else {
				log.Printf("Batch %d processed", batchID)
			}
		}
	}
}

func (tb *TelegramBot) SendMessage(message string) error {
	msg := tgbotapi.NewMessage(tb.conf.ChatID, message)
	_, err := tb.bot.Send(msg)
	return err
}
