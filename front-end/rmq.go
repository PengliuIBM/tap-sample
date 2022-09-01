package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/apimachinery/pkg/util/uuid"
)

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

func connect(clientId string, uri *url.URL) mqtt.Client {
	opts := createClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("ws://%s/ws", uri.Host))
	opts.SetUsername(uri.User.Username())
	//password, _ := uri.User.Password()
	opts.SetPassword("@abc12345D")
	opts.SetClientID(clientId)
	return opts
}

func listen(uri *url.URL, topic string) {
	client := connect("sub", uri)
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))
	})
}

func testChan(ctx context.Context, rid int, cid <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case msg := <-cid:
			fmt.Printf("routine %d receive %d\n", rid, msg)

		case <-ctx.Done():
			fmt.Printf("routine %d receive terminate signal\n", rid)
			return
		}
	}
	fmt.Printf("go route exit")

}

func main() {
	/*
		var wg sync.WaitGroup
		ch := make(chan int)
		ctx, cancelFunc := context.WithCancel(context.Background())
		for thrid := 1; thrid < 6; thrid++ {
			wg.Add(1)
			go testChan(ctx, thrid, ch, &wg)
		}

		for da := 1; da < 30; da++ {
			ch <- da
		}
		cancelFunc()
		wg.Wait()
		return
	*/
	//uri, err := url.Parse(os.Getenv("CLOUDMQTT_URL"))
	msgnub, err := strconv.Atoi(os.Args[1])
	uri, err := url.Parse("amqp://jeffrey:abc12345D@rmq.haas-495.pez.vmware.com:15675/ping")
	//uri := "amqp://jeffrey:abc12345D@rmq.haas-495.pez.vmware.com:15675"
	if err != nil {
		log.Fatal(err)
	}
	topic := uri.Path[1:len(uri.Path)]
	if topic == "" {
		topic = "ping"
	}
	//topic = "alleycat-subscribe"
	topic = "ping"
	client := connect("pub", uri)
	//topic = "mystream"
	for i := 1; i < msgnub+1; i++ {
		ss := Request{
			UUID:            strconv.Itoa(i),
			Event:           "update",
			DeviceTimestamp: int64(i),
			Second:          i,
			RaceId:          int64(i),
			Name:            "jeffrey",
			RacerId:         99,
			ClassId:         3,
			Cadence:         90,
			Resistance:      70}

		r := cloudevents.NewEvent(cloudevents.VersionV1)
		r.SetType("dev.knative.docs.sample")
		r.SetID(string(uuid.NewUUID()))
		r.SetSource("https://github.com/knative/docs/docs/serving/samples/cloudevents/cloudevents-go")
		m, e := json.Marshal(ss)
		if e != nil {
			fmt.Printf("error")
		}
		if err := r.SetData("application/json", m); err != nil {
			fmt.Printf("marshal error")
		}
		sm, e1 := json.Marshal(r)
		if e1 != nil {
			fmt.Printf("error")
		}
		fmt.Println(string(sm))
		client.Publish(topic, 0, false, string(sm))
		time.Sleep(20 * time.Millisecond)
	}

	return

	//go listen(uri, topic)

	// client := connect("pub", uri)
	// timer := time.NewTicker(1 * time.Second)
	// for t := range timer.C {
	// 	client.Publish(topic, 0, false, t.String())
	// }
}
