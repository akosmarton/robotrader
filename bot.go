package main

import (
	"context"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Bot struct {
	startTime time.Time
	roomId    id.RoomID
	client    *mautrix.Client
	msg       chan string
}

func NewBot(homeserver string, userId string, accessToken string, roomId string) *Bot {
	client, err := mautrix.NewClient(homeserver, id.UserID(userId), accessToken)
	if err != nil {
		return nil
	}
	b := &Bot{
		startTime: time.Now(),
		roomId:    id.RoomID(roomId),
		client:    client,
		msg:       make(chan string, 100),
	}
	client.Syncer.(*mautrix.DefaultSyncer).OnEvent(b.handler)
	return b
}

func (b *Bot) handler(ctx context.Context, evt *event.Event) {
	if evt.Timestamp < b.startTime.UnixMilli() {
		return
	}
	if evt.Sender == b.client.UserID {
		return
	}
	if evt.RoomID != b.roomId {
		return
	}
	b.msg <- evt.Content.AsMessage().Body
	b.client.SendReceipt(ctx, evt.RoomID, evt.ID, event.ReceiptTypeRead, mautrix.ReqSetReadMarkers{FullyRead: evt.ID})
}

func (b *Bot) Run(ctx context.Context) error {
	return b.client.SyncWithContext(ctx)
}

func (b *Bot) Message() chan string {
	return b.msg
}

func (b *Bot) SendText(msg string) {
	b.client.SendText(context.Background(), b.roomId, msg)
}

func (b *Bot) SendHtml(msg string) {
	b.client.SendMessageEvent(context.Background(), b.roomId, event.EventMessage, event.MessageEventContent{
		MsgType:       event.MsgText,
		Body:          "",
		Format:        "org.matrix.custom.html",
		FormattedBody: msg,
	})
}

func (b *Bot) SendCode(msg string) {
	b.client.SendMessageEvent(context.Background(), b.roomId, event.EventMessage, event.MessageEventContent{
		MsgType:       event.MsgText,
		Body:          msg,
		Format:        "org.matrix.custom.html",
		FormattedBody: "<pre><code class=\"language-plaintext\">" + msg + "</code></pre>",
	})
}
