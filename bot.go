package main

import (
	"bytes"
	"context"
	"html"
	"strconv"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Bot struct {
	startTime time.Time
	chatID    int64
	client    *bot.Bot
	msg       chan string
}

func (b *Bot) Write(p []byte) (n int, err error) {
	b.SendText(string(p))
	return len(p), nil
}

func NewBot(botToken string, chatId string) *Bot {
	chatID, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		return nil
	}

	b := &Bot{
		startTime: time.Now(),
		chatID:    chatID,
		msg:       make(chan string, 100),
	}

	client, err := bot.New(botToken, bot.WithDefaultHandler(b.handler))
	if err != nil {
		return nil
	}
	b.client = client

	return b
}

func (b *Bot) handler(ctx context.Context, tg *bot.Bot, update *models.Update) {
	_ = ctx
	_ = tg

	if update.Message == nil {
		return
	}
	if update.Message.Chat.ID != b.chatID {
		return
	}
	if int64(update.Message.Date)*1000 < b.startTime.UnixMilli() {
		return
	}

	text := update.Message.Text
	if text == "" {
		text = update.Message.Caption
	}
	if text == "" {
		return
	}

	b.msg <- text
}

func (b *Bot) Run(ctx context.Context) error {
	b.client.Start(ctx)
	return ctx.Err()
}

func (b *Bot) Message() chan string {
	return b.msg
}

func (b *Bot) SendText(msg string) {
	_, _ = b.client.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: b.chatID,
		Text:   msg,
	})
}

func (b *Bot) SendHtml(msg string) {
	_, _ = b.client.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:    b.chatID,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
	})
}

func (b *Bot) SendCode(msg string) {
	b.SendHtml("<pre><code>" + html.EscapeString(msg) + "</code></pre>")
}

func (b *Bot) SendImage(buf []byte, contentType string, filename string) error {
	_ = contentType

	_, err := b.client.SendPhoto(context.Background(), &bot.SendPhotoParams{
		ChatID: b.chatID,
		Photo: &models.InputFileUpload{
			Filename: filename,
			Data:     bytes.NewReader(buf),
		},
		Caption: filename,
	})
	return err
}
