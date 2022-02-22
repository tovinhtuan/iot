package mqtt_go

import (
	"encoding/json"
	"fmt"
	"iot/db/repository"
	"iot/db/storage"
	"iot/model"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	cron "gopkg.in/robfig/cron.v2"
)

var SensorData model.Sensor
var NewData model.Sensor
var PsqlDB *storage.PSQLManager
var SensorRepository repository.SensorRepository
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}
var (
	broker = "broker.emqx.io"
	port   = 1883
	// username = "emqx"
	username = "totuan"
	// clientId = "go_mqtt_client"
	clientId = "go_mqtt_client_iot"
	password = "public"
	// password = "110298"
	topic = "test100"
)

func ConnectMQTTBroker(opts *mqtt.ClientOptions) (*mqtt.Client, error) {
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(clientId)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	// if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
	// 	fmt.Println(token.Error())
	// 	os.Exit(1)
	// }
	return &client, nil
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	// type Sensor struct {
	// 	Flight      bool    `json:"flight"`
	// 	Temperature float64 `json:"temperature"`
	// 	Humidity    float64 `json:"humidity"`
	// 	CreatedAt   time.Time
	// }
	// var SensorData Sensor
	// check := make(chan int, 1)
	err := json.Unmarshal(msg.Payload(), &NewData)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	SensorData = NewData
	// t, _ := time.Parse("2006-01-02 15:04:05", NewData.CreatedAt.Format("2006-01-02 15:04:05"))
	dataDB := model.Sensor{
		Flight:      NewData.Flight,
		Temperature: NewData.Temperature,
		Humidity:    NewData.Humidity,
		CreatedAt:   NewData.CreatedAt,
	}
	err = SensorRepository.InsertSensor(&dataDB)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
	}
}

func Sub(client mqtt.Client) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)
}
func Publish(c *cron.Cron, client mqtt.Client) {
	test := make(chan int)
	local, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	c.AddFunc("@every 0h0m4s", func() {
		t, _ := time.Parse("2006-01-02 15:04:05", time.Now().In(local).Format("2006-01-02 15:04:05"))
		num := 1
		for i := 0; i < num; i++ {
			temperature := fmt.Sprintf("%.2f", -10.0+rand.Float64()*(100.0-(-10.0)))
			humidity := fmt.Sprintf("%.2f", 0.0+rand.Float64()*(100.0-0))
			temfl, _ := strconv.ParseFloat(temperature, 64)
			humfl, _ := strconv.ParseFloat(humidity, 64)
			sensor := model.Sensor{
				Flight:      false,
				Temperature: temfl,
				Humidity:    humfl,
				CreatedAt:   t,
			}
			messageJSON, err := json.Marshal(sensor)
			if err != nil {
				log.Printf("Error marshal sensor:%v\n", err)
				os.Exit(1)
			}
			token := client.Publish(topic, 0, false, messageJSON)
			token.Wait()
			time.Sleep(time.Second * 10)
		}
	})
	c.Start()
	<-test
}
