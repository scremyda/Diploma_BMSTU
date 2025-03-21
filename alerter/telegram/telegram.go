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

func NewTelegramBot(repo repo.QueueRepo, conf Config) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(conf.BotToken)
	if err != nil {
		return nil, fmt.Errorf("error creating telegram bot: %w", err)
	}

	return &TelegramBot{
		repo: repo,
		bot:  bot,
		conf: conf,
	}, nil
}

func (tb *TelegramBot) SendMessage(message string) error {
	msg := tgbotapi.NewMessage(tb.conf.ChatID, message)
	_, err := tb.bot.Send(msg)
	return err
}

func (tb *TelegramBot) ProcessQueue(ctx context.Context, pollInterval time.Duration) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stop events processing")
			return
		default:
		}

		batchID, err := tb.repo.NextBatch(ctx)
		if err != nil {
			log.Printf("Error receiving new batch: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		if batchID <= 0 {
			// Нет новых событий – ждем указанное время
			time.Sleep(pollInterval)
			continue
		}

		events, err := tb.repo.GetBatchEvents(ctx, batchID)
		if err != nil {
			log.Printf("Error getting batch events for batch %d: %v", batchID, err)
			tb.repo.FinishBatch(ctx, batchID)
			time.Sleep(pollInterval)
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
