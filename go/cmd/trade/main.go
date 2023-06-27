package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rodrigoschaer/homebroker/go/internal/infra/kafka"
	"github.com/rodrigoschaer/homebroker/go/internal/market/dto"
	"github.com/rodrigoschaer/homebroker/go/internal/market/entity"
	"github.com/rodrigoschaer/homebroker/go/internal/market/transformer"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() { //Thread1
	ordersIn := make(chan *entity.Order)
	ordersOut := make(chan *entity.Order)
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	kafkaMsgChan := make(chan *ckafka.Message)
	configMap := &ckafka.ConfigMap{
		"bootstrap.servers": "host.docker.internal:9094",
		"group.id":          "myGroup",
		"auto.offset.reset": "latest",
	}

	producer := kafka.NewProducer(configMap)
	kafka := kafka.NewConsumer(configMap, []string{"input-orders"})

	go kafka.Consume(kafkaMsgChan) //Thread2

	book := entity.NewBook(ordersIn, ordersOut, wg)
	go book.Trade() //Thread3

	go func() { //Thread4
		for msg := range kafkaMsgChan {
			wg.Add(1)
			tradeInput := dto.TradeInput{}
			err := json.Unmarshal(msg.Value, &tradeInput)
			if err != nil {
				panic(err)
			}
			order := transformer.TransformInput(tradeInput)
			ordersIn <- order
		}
	}()

	for res := range ordersOut {
		output := transformer.TransformOutput(res)
		outputJson, err := json.Marshal(output)
		if err != nil {
			fmt.Println(err)
		}
		producer.Publish(outputJson, []byte("orders"), "output")
	}

}