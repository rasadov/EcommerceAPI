package main

import (
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/rasadov/EcommerceAPI/product/config"
	"github.com/tinrab/retry"

	"github.com/rasadov/EcommerceAPI/product/internal"
)

func main() {
	var repository internal.Repository

	var producer sarama.AsyncProducer
	producer, err := sarama.NewAsyncProducer([]string{config.BootstrapServers}, nil)
	if err != nil {
		log.Println("Kafka producer unavailable:", err)
		producer = nil
	} else {
		defer func() {
			if err := producer.Close(); err != nil {
				log.Println(err)
			}
		}()
	}

	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		repository, err = internal.NewElasticRepository(config.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})
	defer repository.Close()
	log.Println("Listening on port 8080...")
	service := internal.NewProductService(repository, producer)
	log.Fatal(internal.ListenGRPC(service, 8080))
}
