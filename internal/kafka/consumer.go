package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"myapp/internal/model"
	"myapp/internal/service"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Consumer struct {
	consumer *kafka.Consumer
	service  service.Service
	topic    string
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
		consumer: consumer,
		service:  service,
		topic:    topic,
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
					if err.(kafka.Error).Code() == kafka.ErrTimedOut {
						continue
					}
					log.Printf("Consumer error: %v", err)
					continue
				}

				log.Printf("Received message from Kafka: %s", string(msg.Value))
				if err := c.processMessage(msg); err != nil {
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

func (c *Consumer) processMessage(msg *kafka.Message) error {
	log.Printf("Received message: %s", string(msg.Value))

	if len(msg.Value) == 0 {
		log.Printf("Empty message received, skipping")
		return nil
	}

	var order model.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return nil
	}

	log.Printf("Successfully unmarshaled order: %s", order.OrderUID)

	if order.OrderUID == "" {
		log.Printf("Order UID is empty, skipping message")
		return nil
	}

	if err := c.service.ProcessOrder(&order); err != nil {
		log.Printf("Failed to process order %s: %v", order.OrderUID, err)
		return nil
	}

	log.Printf("Successfully processed order: %s", order.OrderUID)
	return nil
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
