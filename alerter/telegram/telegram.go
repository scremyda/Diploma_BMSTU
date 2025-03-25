package telegram

import (
	"context"
	"diploma/alerter/consumer"
	"errors"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	BotToken         string        `yaml:"bot_token"`
	ChatID           int64         `yaml:"chat_id"`
	MessagesInterval time.Duration `yaml:"messages_interval"`
}

type TelegramBot struct {
	consumer consumer.Interface
	bot      *tgbotapi.BotAPI
	conf     Config
}

func NewTelegramBot(consumer consumer.Interface, conf Config) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(conf.BotToken)
	if err != nil {
		return nil, fmt.Errorf("error creating telegram bot: %w", err)
	}

	return &TelegramBot{
		consumer: consumer,
		bot:      bot,
		conf:     conf,
	}, nil
}

func (tb *TelegramBot) ProcessMessages(ctx context.Context) {
	ticker := time.NewTicker(tb.conf.MessagesInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stop events processing")
			return
		case <-ticker.C:
			events, err := tb.consumer.GetEvents(ctx)
			if err != nil {
				if !errors.Is(err, consumer.ErrFinishBatch) {
					log.Printf("Error getting events: %v", err)
					continue
				}
			}

			for _, event := range events {
				text := fmt.Sprintf("Error for event %s: %s", event.Target, event.Message)
				if err := tb.SendMessage(text); err != nil {
					log.Printf("Error sending message to Telegram: %v", err)
				} else {
					log.Printf("Message sent: %s", text)
				}
			}
		}
	}
}

func (tb *TelegramBot) SendMessage(message string) error {
	msg := tgbotapi.NewMessage(tb.conf.ChatID, message)
	_, err := tb.bot.Send(msg)
	return err
}
