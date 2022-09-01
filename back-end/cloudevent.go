package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kelseyhightower/envconfig"
)

// Request is the structure of the event we expect to receive.
type Request struct {
	UUID            string  `json:"uuid"`
	Event           string  `json:"event"` //
	DeviceTimestamp int64   `json:"deviceTimestamp"`
	Second          int     `json:"second"`
	RaceId          int64   `json:"raceId"`
	Name            string  `json:"name"`
	RacerId         int     `json:"racerId"`
	ClassId         int     `json:"classId"`
	Cadence         float32 `json:"cadence"`
	Resistance      float32 `json:"resistance"`
}

// Response is the structure of the event we send in response to requests.
type Response struct {
	Message string `json:"message,omitempty"`
}

type Receiver struct {
	client cloudevents.Client
	ch     chan<- Request
	//mqttClient mqtt.Client
	// If the K_SINK environment variable is set, then events are sent there,
	// otherwise we simply reply to the inbound request.
	Target string `envconfig:"K_SINK"`
}

type RealtimeStatus struct {
	Msg     string          `json:"msg"`
	ClassId int             `json:"classId"`
	RaceId  int64           `json:"raceId"`
	Ts      int64           `json:"ts"`
	Results map[int]float64 `json:"result"`
}

func connect2rmq() (mqtt.Client, error) {

	rabbitmq_url := os.Getenv("RABBITMQ_URL")
	rabbitmq_user := os.Getenv("RABBITMQ_USER")
	rabbitmq_password := os.Getenv("RABBITMQ_PASSWORD")
	rabbitmq_topic := os.Getenv("RABBITMQ_TOPIC")
	log.Printf("Rabbitmq config: url-%q; usename-%q; topic-%q", rabbitmq_url, rabbitmq_user, rabbitmq_topic)

	/*
		rabbitmq_url := "rmq.haas-495.pez.vmware.com:15675"
		rabbitmq_user := "jeffrey"
		rabbitmq_password := "@abc12345D"
		rabbitmq_topic := "mystream"
	*/
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("ws://%s/ws", rabbitmq_url))
	opts.SetUsername(rabbitmq_user)
	opts.SetPassword(rabbitmq_password)
	opts.SetClientID("Streamsvr")

	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return client, nil
}

func formatmsg(msg Request) string {
	m, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	rs := new(RealtimeStatus)
	rs.ClassId = msg.ClassId
	rs.Ts = msg.DeviceTimestamp
	rs.RaceId = msg.RaceId
	rs.Results = make(map[int]float64)
	rs.Results[msg.RacerId] = math.Round(((float64(msg.Cadence) + 35.00) * (float64(msg.Resistance) + 65.00)) / 100.00)
	rs.Msg = "results"

	m, e := json.Marshal(rs)
	if e != nil {
		log.Fatal(e)
		return ""
	}
	return string(m)
}

func publish2mq(req <-chan Request) {

	rabbitmq_url := os.Getenv("RABBITMQ_URL")
	rabbitmq_user := os.Getenv("RABBITMQ_USER")
	rabbitmq_password := os.Getenv("RABBITMQ_PASSWORD")
	rabbitmq_topic := os.Getenv("RABBITMQ_TOPIC")
	log.Printf("Rabbitmq config: url-%q; usename-%q; topic-%q", rabbitmq_url, rabbitmq_user, rabbitmq_topic)

	/*
		rabbitmq_url := "rmq.haas-495.pez.vmware.com:15675"
		rabbitmq_user := "jeffrey"
		rabbitmq_password := "@abc12345D"
		rabbitmq_topic := "mystream"
	*/
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("ws://%s/ws", rabbitmq_url))
	opts.SetUsername(rabbitmq_user)
	opts.SetPassword(rabbitmq_password)
	opts.SetClientID("Streamsvr")

	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
		return
	}

	for {
		msg := <-req
		log.Printf("Receive update event: %q", msg.Name)
		var sendmsg string
		if msg.Event == "update" {
			m, err := json.Marshal(msg)
			if err != nil {
				log.Fatal(err)
				continue
			}
			rs := new(RealtimeStatus)
			rs.ClassId = msg.ClassId
			rs.Ts = msg.DeviceTimestamp
			rs.RaceId = msg.RaceId
			rs.Results = make(map[int]float64)
			rs.Results[msg.RacerId] = math.Round(((float64(msg.Cadence) + 35.00) * (float64(msg.Resistance) + 65.00)) / 100.00)
			rs.Msg = "results"

			m, e := json.Marshal(rs)
			if e != nil {
				log.Fatal(e)
				continue
			}
			sendmsg = string(m)
		} else if msg.Event == "final" {
			log.Printf("Receive update event: %q", msg.Name)
		}

		log.Printf("send message: %q", sendmsg)
		client.Publish(rabbitmq_topic, 0, false, sendmsg)

		//if tok := client.Publish(rabbitmq_topic, 0, false, sendmsg); !tok.WaitTimeout(2) {
		//	log.Fatal("Publish message timeout")
		//}
	}

}

func main() {
	client, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatal(err.Error())
	}

	//mqttclient, err := connect2rmq()
	//if err != nil {
	//	log.Fatal("Cant' connect to rabbitmq ")
	//}
	c := make(chan Request)
	go publish2mq(c)
	/*
		for i := 1; i < 10; i++ {
			teststruct := Request{
				UUID:            "652f68ea-26f4-46aa-9698-83068d41d6fe",
				Event:           "update",
				DeviceTimestamp: 1633948145189,
				Second:          i,
				raceId:          1,
				Name:            "Jeffrey",
				RacerId:         1633942419496,
				ClassId:         1,
				Cadence:         79.9,
				Resistance:      78}
			c <- teststruct
		}
	*/
	r := Receiver{client: client, ch: c}
	if err := envconfig.Process("", &r); err != nil {
		log.Fatal(err.Error())
	}

	// Depending on whether targeting data has been supplied,
	// we will either reply with our response or send it on to
	// an event sink.
	var receiver interface{} // the SDK reflects on the signature.
	if r.Target == "" {
		receiver = r.ReceiveAndReply
	} else {
		receiver = r.ReceiveAndSend
	}

	if err := client.StartReceiver(context.Background(), receiver); err != nil {
		log.Fatal(err)
	}
}

// handle shared the logic for producing the Response event from the Request.
func handle(req Request) Response {
	return Response{Message: fmt.Sprintf("Hello, %s", req.Name)}
}

// ReceiveAndSend is invoked whenever we receive an event.
func (recv *Receiver) ReceiveAndSend(ctx context.Context, event cloudevents.Event) cloudevents.Result {
	req := Request{}
	if err := event.DataAs(&req); err != nil {
		return cloudevents.NewHTTPResult(400, "failed to convert data: %s", err)
	}
	log.Printf("Got an event from: %q", req.Name)

	resp := handle(req)
	log.Printf("Sending event: %q", resp.Message)

	r := cloudevents.NewEvent(cloudevents.VersionV1)
	r.SetType("dev.knative.docs.sample")
	r.SetSource("https://github.com/knative/docs/docs/serving/samples/cloudevents/cloudevents-go")
	if err := r.SetData("application/json", resp); err != nil {
		return cloudevents.NewHTTPResult(500, "failed to set response data: %s", err)
	}

	ctx = cloudevents.ContextWithTarget(ctx, recv.Target)
	return recv.client.Send(ctx, r)
}

// ReceiveAndReply is invoked whenever we receive an event.
func (recv *Receiver) ReceiveAndReply(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, cloudevents.Result) {
	req := Request{}
	if err := event.DataAs(&req); err != nil {
		log.Printf("failed to convert data: %s", err)
		return nil, cloudevents.NewHTTPResult(400, "failed to convert data: %s", err)
	}
	log.Printf("Got an event from: %q", req.Name)

	//msg := formatmsg(req)
	//log.Print(msg)
	//recv.mqttClient.Publish("rtinto", 0, false, msg)
	recv.ch <- req
	return nil, nil
	/*
		resp := handle(req)
		log.Printf("Replying with event: %q", resp.Message)

		r := cloudevents.NewEvent(cloudevents.VersionV1)
		r.SetType("dev.knative.docs.sample")
		r.SetSource("https://github.com/knative/docs/docs/serving/samples/cloudevents/cloudevents-go")
		if err := r.SetData("application/json", resp); err != nil {
			return nil, cloudevents.NewHTTPResult(500, "failed to set response data: %s", err)
		}
		return &r, nil
	*/
}
