package kafka

import (
	"context"
	"fmt"
	"github.com/ikondratev/event-service/internal/settings"

	kafkago "github.com/segmentio/kafka-go"
)
type Publisher interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
	Close() error
}

type Producer struct {
	writer *kafkago.Writer
}

func NewProducer(settings *settings.Settings) *Producer {
	return &Producer{
		writer: &kafkago.Writer{
			Addr: kafkago.TCP(settings.Kafka.Brokers...),
			Balancer: &kafkago.LeastBytes{},
		},
	}
}

func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte) error {
	err := p.writer.WriteMessages(ctx, kafkago.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("Kafka publish erorr: %w", err)
	}

	return nil
}

func (p *Producer) Flush(ctx context.Context) error {
	return p.writer.WriteMessages(ctx)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}