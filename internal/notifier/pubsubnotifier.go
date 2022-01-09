package notifier

import (
	"context"

	"gocloud.dev/pubsub"
)

type PubSubNotifier struct {
	topic pubsub.Topic
}

func NewPubSubNotifier(topic pubsub.Topic) *PubSubNotifier {
	return &PubSubNotifier{
		topic: topic,
	}
}

func (p *PubSubNotifier) NotifySubscriber(ctx context.Context, msg string) error {
	err := p.topic.Send(ctx, &pubsub.Message{
		Body: []byte(msg),
	})
	return err
}
