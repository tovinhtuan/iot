package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"iot/db/repository"
	"iot/db/repository/mqtt_go"
	"iot/db/storage"
	"iot/model"
	"log"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	cron "gopkg.in/robfig/cron.v2"
)
type Broker struct {

	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool
}
func NewServer() (broker *Broker) {
	// Instantiate a broker
	broker = &Broker{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
	}

	// Set it running - listening and broadcasting events
	go broker.listen()

	return
}
var data model.Sensor

func (broker *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseGlob("hello.html")
	// Make sure that the writer supports flushing.
	//
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.closingClients <- messageChan
	}()

	// Listen to connection close and un-register messageChan
	// notify := rw.(http.CloseNotifier).CloseNotify()
	notify := req.Context().Done()

	go func() {
		<-notify
		broker.closingClients <- messageChan
	}()

	for {

		// Write to the ResponseWriter
		// Server Sent Events compatible
		// fmt.Fprintf(rw, "data: %s\n\n", <-messageChan)
		
		_ = json.Unmarshal(<-messageChan, &data)
		dataSend := getDataRequest(data)
		t.Execute(rw, map[string]template.HTML{"flight": template.HTML(dataSend[0]), "humidity": template.HTML(dataSend[1]), "temperature": template.HTML(dataSend[2]), "date": template.HTML(dataSend[3])})
		// t.Execute(rw, data)
		// Flush the data immediatly instead of buffering it for later.
		flusher.Flush()
	}

}
func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.newClients:

			// A new client has connected.
			// Register their message channel
			broker.clients[s] = true
			log.Printf("Client added. %d registered clients", len(broker.clients))
		case s := <-broker.closingClients:

			// A client has dettached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
			log.Printf("Removed client. %d registered clients", len(broker.clients))
		case event := <-broker.Notifier:

			// We got a new event from the outside!
			// Send event to all connected clients
			for clientMessageChan, _ := range broker.clients {
				clientMessageChan <- event
			}
		}
	}

}
func getDataRequest(data model.Sensor) []string {
	var dataReceive []string
	dataReceive = append(dataReceive, fmt.Sprintf("%v", data.Flight), fmt.Sprintf("%v", data.Humidity), 
	fmt.Sprintf("%v", data.Temperature), fmt.Sprintf("%v", data.CreatedAt.Format("2006-01-02 15:04:05")))
	return dataReceive
}
func main() {
	//Connect postgresql
	var err error
	mqtt_go.PsqlDB, err = storage.NewPSQLManager()
	if err != nil {
		log.Fatalf("Error when connecting database, err: %v", err)
	}
	defer mqtt_go.PsqlDB.Close()
	mqtt_go.SensorRepository = repository.NewSensorRepository(mqtt_go.PsqlDB)

	cr := cron.New()
	//Create mqtt client
	opts := mqtt.NewClientOptions()
	client, err := mqtt_go.ConnectMQTTBroker(opts)
	if err != nil {
		log.Printf("Error connect broker: %v\n", err)
	}
	go func() {
		mqtt_go.Sub(*client)
		mqtt_go.Publish(cr, *client)
		(*client).Disconnect(2500)
		cr.Stop()
	}()
	// var data *model.Sensor
	// then := time.Now().Add(time.Duration(-10) * time.Second).Format("2006-01-02 15:04")
	// currentTime, _ := time.Parse("2006-01-02 15:04", then)
	// time.Sleep(time.Second*10)
	// data, err = mqtt_go.SensorRepository.GetSensorByTime(currentTime)
	// if err != nil {
	// 	log.Printf("Error when get sensor database, err: %v", err)
	// }
	// fmt.Println("day la msg:", mqtt_go.SensorData)
	// fmt.Printf("data: %v\n", data)
	// http.HandleFunc("/send", httpHandler)
	broker := NewServer()

	go func() {
		for {
			time.Sleep(time.Second * 4)
			// eventString := fmt.Sprintf("%v", mqtt_go.SensorData)
			dataJSON, _ := json.Marshal(mqtt_go.SensorData)
			log.Println("Receiving event")
			broker.Notifier <- dataJSON
		}
	}()

	log.Fatal("HTTP server error: ", http.ListenAndServe("localhost:3000", broker))
}
