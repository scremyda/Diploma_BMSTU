package telegram

import (
	"context"
	"diploma/alerter/repo"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   int64  `yaml:"chat_id"`
}

type TelegramBot struct {
	repo         repo.QueueRepo
	bot          *tgbotapi.BotAPI
	chatID       int64
	queueName    string
	consumerName string
}

func NewTelegramBot(repo repo.QueueRepo, token string, chatID int64, queueName, consumerName string) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания Telegram бота: %w", err)
	}

	return &TelegramBot{
		repo:         repo,
		bot:          bot,
		chatID:       chatID,
		queueName:    queueName,
		consumerName: consumerName,
	}, nil
}

func (tb *TelegramBot) SendMessage(message string) error {
	msg := tgbotapi.NewMessage(tb.chatID, message)
	_, err := tb.bot.Send(msg)
	return err
}

func (tb *TelegramBot) ProcessQueue(ctx context.Context, pollInterval time.Duration) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка обработки очереди в Telegram")
			return
		default:
		}

		batchID, err := tb.repo.NextBatch(ctx, tb.queueName, tb.consumerName)
		if err != nil {
			log.Printf("Ошибка получения следующей batch: %v", err)
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
			log.Printf("Ошибка получения событий batch %d: %v", batchID, err)
			tb.repo.FinishBatch(ctx, tb.queueName, tb.consumerName, batchID)
			time.Sleep(pollInterval)
			continue
		}

		for _, event := range events {
			text := fmt.Sprintf("Ошибка для %s: %s", event.Target, event.Message)
			if err := tb.SendMessage(text); err != nil {
				log.Printf("Ошибка отправки сообщения в Telegram: %v", err)
			} else {
				log.Printf("Отправлено сообщение: %s", text)
			}
		}

		if err := tb.repo.FinishBatch(ctx, tb.queueName, tb.consumerName, batchID); err != nil {
			log.Printf("Ошибка завершения batch %d: %v", batchID, err)
		} else {
			log.Printf("Batch %d обработана", batchID)
		}
	}
}
