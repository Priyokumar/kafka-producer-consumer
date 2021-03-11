package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	kafka "github.com/segmentio/kafka-go"
)

var (
	messageCount int
	kafkaURL     string
	topic        string
)

func getKafkaReader(kafkaURL, topic, groupID string) *kafka.Reader {
	brokers := strings.Split(kafkaURL, ",")
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}

func main() {

	// get kafka reader using environment variables.
	kafkaURL = os.Getenv("kafkaURL")
	topic = os.Getenv("topic")
	log.Println("Reading environment variables. kafkaURL : " + kafkaURL + ", topic : " + topic)

	// Handling Swagger UI at the base path
	swaagerUIFs := http.FileServer(http.Dir("./swagger-ui"))
	http.Handle("/", swaagerUIFs)

	apiBasePath := "/v1/api"
	port := ":8000"
	/// Handling message count
	http.HandleFunc(apiBasePath+"/messages/count", handleMessageCountFunc)

	go startConsuming("g-1")
	go startConsuming("g-2")

	// Running the web server.
	log.Println("Started consumer service on port " + port)
	log.Fatal(http.ListenAndServe(port, nil))
	//routeEngine.Run()
}

func handleMessageCountFunc(wrt http.ResponseWriter, req *http.Request) {
	JSON(200, messageCount, wrt)
}

func startConsuming(groupID string) {

	reader := getKafkaReader(kafkaURL, topic, groupID)
	mutex := &sync.Mutex{}
	defer reader.Close()
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("message at topic:%v partition:%v offset:%v	%s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
		mutex.Lock()
		messageCount++
		fmt.Println(messageCount)
		mutex.Unlock()
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
