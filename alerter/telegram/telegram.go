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
			alerterEvents, err := tb.consumer.GetAlerterEvents(ctx)
			if err != nil {
				if !errors.Is(err, consumer.ErrFinishBatch) {
					log.Printf("Error getting alerter events: %v", err)
					continue
				}
			}

			for _, alerterEvent := range alerterEvents {
				text := fmt.Sprintf("Сообщение alerter для %s\n\n%s", alerterEvent.Target, alerterEvent.Message)
				if err := tb.SendMessage(text); err != nil {
					log.Printf("Error sending message to Telegram: %v", err)
				} else {
					log.Printf("Message alerter sent: %s", text)
				}
			}

			//certerEvents, err := tb.consumer.GetCerterEvents(ctx)
			//if err != nil {
			//	if !errors.Is(err, consumer.ErrFinishBatch) {
			//		log.Printf("Error getting certer events: %v", err)
			//		continue
			//	}
			//}
			//
			//for _, certerEvent := range certerEvents {
			//	text := fmt.Sprintf("Сообщение certer: %s", certerEvent.Target)
			//	if err := tb.SendMessage(text); err != nil {
			//		log.Printf("Error sending message to Telegram: %v", err)
			//	} else {
			//		log.Printf("Message certer sent: %s", text)
			//	}
			//}
		}
	}
}

func (tb *TelegramBot) SendMessage(message string) error {
	msg := tgbotapi.NewMessage(tb.conf.ChatID, message)
	_, err := tb.bot.Send(msg)
	return err
}
