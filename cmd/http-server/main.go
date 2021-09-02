package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/maestre3d/stream-api/controller"
	"github.com/maestre3d/stream-api/internal/event"
	"github.com/maestre3d/stream-api/internal/persistence"
	"github.com/neutrinocorp/gluon"
	_ "github.com/neutrinocorp/gluon/gkafka"
)

func main() {
	r := mux.NewRouter()
	b := initBus()
	initRoutes(r, b)
	ws := initStreamAPI(r, b)
	defer ws.Close()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := b.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	panic(srv.ListenAndServe())
}

func initRoutes(r *mux.Router, b *gluon.Bus) {
	repo := persistence.NewOrderInMemory()
	orderCtrl := controller.NewOrderHttp(repo, b)
	orderCtrl.MapRoutes(r)
}

func initStreamAPI(r *mux.Router, b *gluon.Bus) *controller.OrderStreamsWs {
	orderStreams := controller.NewOrderStreamsWs(b)
	orderStreams.MapStreams(r)
	return orderStreams
}

func initBus() *gluon.Bus {
	b := gluon.NewBus("kafka",
		gluon.WithCluster("localhost:9092", "localhost:9093", "localhost:9094"))
	b.RegisterSchema(event.OrderIssued{},
		gluon.WithTopic("org.neutrino.marketplace.event.order.issued"))
	b.RegisterSchema(event.OrderUpdated{},
		gluon.WithTopic("org.neutrino.marketplace.event.order.updated"))

	b.Subscribe(event.OrderIssued{}).HandlerFunc(func(ctx context.Context, msg *gluon.Message) error {
		log.Printf("order received: %+v", msg.Data)
		return nil
	})
	b.Subscribe(event.OrderUpdated{}).HandlerFunc(func(ctx context.Context, msg *gluon.Message) error {
		log.Printf("order updated: %+v", msg.Data)
		return nil
	})
	return b
}
