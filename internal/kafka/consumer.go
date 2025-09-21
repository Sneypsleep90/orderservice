package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"myapp/internal/model"
	"myapp/internal/service"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type Consumer struct {
	consumer   *kafka.Consumer
	service    service.Service
	topic      string
	dlq        *Producer
	maxRetries int
}

func NewConsumer(brokers, groupID, topic string, service service.Service) (*Consumer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": true,
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &Consumer{
		consumer:   consumer,
		service:    service,
		topic:      topic,
		maxRetries: 3,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("Starting Kafka consumer for topic: %s", c.topic)

	if err := c.consumer.Subscribe(c.topic, nil); err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	log.Println("Kafka consumer started and subscribed to topic")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
				if err != nil {
					var ke kafka.Error
					if errors.As(err, &ke) && ke.Code() == kafka.ErrTimedOut {
						continue
					}
					log.Printf("Consumer error: %v", err)
					continue
				}

				log.Printf("Received message from Kafka: %s", string(msg.Value))
				if err := c.processWithRetry(msg); err != nil {
					log.Printf("Error processing message: %v", err)
				}
			}
		}
	}()

	return nil
}

func (c *Consumer) Stop() error {
	return c.consumer.Close()
}

func (c *Consumer) SetDLQProducer(p *Producer) {
	c.dlq = p
}

func (c *Consumer) processMessage(msg *kafka.Message) error {
	tracer := otel.Tracer("kafka")
	ctx, span := tracer.Start(context.TODO(), "processMessage")
	defer span.End()
	log.Printf("Received message: %s", string(msg.Value))

	if len(msg.Value) == 0 {
		log.Printf("Empty message received, skipping")
		return nil
	}

	var order model.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		span.SetStatus(codes.Error, err.Error())
		return nil
	}

	log.Printf("Successfully unmarshaled order: %s", order.OrderUID)

	if order.OrderUID == "" {
		log.Printf("Order UID is empty, skipping message")
		return nil
	}

	if err := c.service.ProcessOrder(&order); err != nil {
		log.Printf("Failed to process order %s: %v", order.OrderUID, err)
		span.SetStatus(codes.Error, err.Error())
		return nil
	}

	log.Printf("Successfully processed order: %s", order.OrderUID)
	_ = ctx
	return nil
}

func (c *Consumer) processWithRetry(msg *kafka.Message) error {
	var err error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if err = c.processMessage(msg); err == nil {
			return nil
		}
		time.Sleep(time.Duration(200*(attempt+1)) * time.Millisecond)
	}
	if c.dlq != nil {
		// Send to DLQ
		if dlqErr := c.dlq.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &c.dlq.topic, Partition: kafka.PartitionAny},
			Value:          msg.Value,
			Headers:        msg.Headers,
		}, nil); dlqErr != nil {
			log.Printf("Failed to write message to DLQ: %v", dlqErr)
		} else {
			c.dlq.producer.Flush(5000)
			log.Printf("Message sent to DLQ topic %s", c.dlq.topic)
		}
	}
	return err
}

type Producer struct {
	producer *kafka.Producer
	topic    string
}

func NewProducer(brokers, topic string) (*Producer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": brokers,
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *Producer) SendOrder(order *model.Order) error {
	orderBytes, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          orderBytes,
	}

	if err := p.producer.Produce(msg, nil); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	p.producer.Flush(15 * 1000)
	log.Printf("Message sent to topic %s", p.topic)
	return nil
}

func (p *Producer) Close() error {
	p.producer.Close()
	return nil
}
