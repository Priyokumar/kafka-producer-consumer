package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

var (
	kafkaURL          string
	topic             string
	startedPublishing bool
)

func main() {

	kafkaURL = os.Getenv("kafkaURL")
	topic = os.Getenv("topic")
	log.Println("Read environment variables. kafkaURL : " + kafkaURL + ", topic : " + topic)

	apiBasePath := "/v1/api"
	port := ":8001"
	// Handling Swagger UI
	swaagerUIFs := http.FileServer(http.Dir("./swagger-ui"))
	http.Handle("/", swaagerUIFs)
	// Add handle func for publishing message.
	http.HandleFunc(apiBasePath+"/message/publish", publishMessageHandler)

	// Running the web server.
	log.Println("Started consumer service on port " + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func publishMessageHandler(wrt http.ResponseWriter, req *http.Request) {

	log.Println("Inside publishHandler")

	if startedPublishing == false {
		startedPublishing = true
		go startPublishingMessage()
		JSON(200, "Started publishing message.", wrt)
	} else {
		JSON(200, "Already started publishing message. No need start again. There is no Enpoint to stop publishing message for now. We will implement this later.", wrt)
	}

}

func startPublishingMessage() {

	kafkaWriter := getKafkaWriter(kafkaURL, topic)
	defer kafkaWriter.Close()
	var count int64 = 0

	for {
		message := getMessage(count)

		err := kafkaWriter.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(strconv.FormatInt(count, 36)),
			Value: message,
		})
		if err != nil {

			// it will try upto 3 times
			if count > 3 {
				panic("write failed" + err.Error())
			} else {
				log.Println("write failed" + err.Error())
			}
		}

		fmt.Println("writes:", count)
		count++
		time.Sleep(time.Second)
	}
}

// it will return to two different json data
func getMessage(i int64) []byte {

	data1 := map[string]interface{}{

		"name":    "priyo" + strconv.FormatInt(i, 36),
		"address": "address" + strconv.FormatInt(i, 36),
		"age":     i + 1,
	}

	data2 := map[string]interface{}{

		"empId":        i,
		"salary":       i + 100000,
		"workLocation": "Bangalore",
	}

	d1, _ := json.Marshal(data1)
	d2, _ := json.Marshal(data2)

	if i%2 == 0 {
		return d1
	}
	return d2

}

func getKafkaWriter(kafkaURL, topic string) *kafka.Writer {
	log.Println("Initializing kafka writer.")
	return &kafka.Writer{
		Addr:     kafka.TCP(kafkaURL),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

// JSON Respond json data
func JSON(code int, data interface{}, writer http.ResponseWriter) {

	resp, error := json.Marshal(data)

	if error != nil {
		log.Println("An error was occured while marshalling the data")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	writer.Write(resp)

}
