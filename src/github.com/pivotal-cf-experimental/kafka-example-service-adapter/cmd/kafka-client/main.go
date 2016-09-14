package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/Shopify/sarama.v1"
)

func main() {
	bootstrapServerArg := flag.String("bootstrapServers", "localhost:9092", "comma separated list of kafka brokers to bootstrap with")
	replayArg := flag.Bool("replay", false, "whether to replay old messages")
	flag.Parse()
	bootstrapServers := strings.Split(*bootstrapServerArg, ",")

	action := flag.Args()[0]
	topic := flag.Args()[1]

	config := sarama.NewConfig()
	client, err := sarama.NewClient(bootstrapServers, config)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	switch action {
	case "produce":
		produce(client, topic)
	case "consume":
		consume(client, topic, *replayArg)
	default:
		panic(fmt.Sprintf("action %s not supported", action))
	}
}

func produce(client sarama.Client, topic string) {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	log.Printf("producing %s ...\n", topic)
	for scanner.Scan() {
		line := scanner.Text()
		message := sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(line)}
		producer.SendMessage(&message)
	}
}

func consume(client sarama.Client, topic string, replay bool) {
	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		panic(err)
	}

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		panic(err)
	}

	log.Printf("partitions are %v\n", partitions)

	var startingOffset int64
	if replay {
		startingOffset = sarama.OffsetOldest
	} else {
		startingOffset = sarama.OffsetNewest
	}

	lines := make(chan string)

	for _, partition := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(topic, partition, startingOffset)
		if err != nil {
			panic(err)
		}

		go func() {
			for message := range partitionConsumer.Messages() {
				lines <- string(message.Value)
			}
		}()
	}

	log.Printf("consuming %s ...\n", topic)
	for line := range lines {
		fmt.Println(line)
	}
}
