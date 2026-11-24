package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type QueueClient struct {
	js jetstream.JetStream
}

func NewQueueClient(url string) (*QueueClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	_, _ = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "WEBHOOKS",
		Subjects: []string{"webhooks.events"},
	})

	return &QueueClient{js: js}, nil
}

func (q *QueueClient) Publish(ctx context.Context, eventID string) error {
	_, err := q.js.Publish(ctx, "webhooks.events", []byte(eventID))
	return err
}

func (q *QueueClient) Consume(ctx context.Context, handler func(eventID string) error) error {
	cons, err := q.js.CreateOrUpdateConsumer(ctx, "WEBHOOKS", jetstream.ConsumerConfig{
		Durable: "WebhookWorker",
	})
	if err != nil {
		return err
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		eventID := string(msg.Data())
		if err := handler(eventID); err == nil {
			msg.Ack()
		} else {
			msg.Nak()
		}
	})
	if err != nil {
		return err
	}

	<-ctx.Done()
	cc.Stop()
	return nil
}
