package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

func ConnectProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	return sarama.NewSyncProducer(brokers, config)
}

func PushOrderToKafka(producer *sarama.SyncProducer, topic string, msg []byte) error {

	// Create a new message
	me := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(msg),
	}

	// Send message

	partition, offset, err := (*producer).SendMessage(me)
	if err != nil {
		return err
	}

	log.Printf("Order is stored in topic(%s)/partition(%d)/offset(%d)\n",
		topic,
		partition,
		offset)

	return nil
}
