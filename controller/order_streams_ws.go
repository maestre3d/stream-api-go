package controller

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/maestre3d/stream-api/internal/event"
	"github.com/neutrinocorp/gluon"
)

type WebSocket interface {
	MapStreams(r *mux.Router)
	Close()
}

type OrderStreamsWs struct {
	bus       *gluon.Bus
	orderChan chan event.OrderUpdated
}

var _ WebSocket = &OrderStreamsWs{}

func NewOrderStreamsWs(b *gluon.Bus) *OrderStreamsWs {
	return &OrderStreamsWs{
		bus:       b,
		orderChan: make(chan event.OrderUpdated),
	}
}

func (o *OrderStreamsWs) MapStreams(r *mux.Router) {
	o.bus.Subscribe(event.OrderUpdated{}).HandlerFunc(func(ctx context.Context, msg *gluon.Message) error {
		o.orderChan <- msg.Data.(event.OrderUpdated)
		return nil
	})
	r.Path("/streams/orders/{id}").Methods(http.MethodGet).HandlerFunc(o.listenToUpdates)
}

func (o *OrderStreamsWs) Close() {
	close(o.orderChan)
}

var upgrader = websocket.Upgrader{} // use default options

func (o *OrderStreamsWs) listenToUpdates(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		select {
		case order := <-o.orderChan:
			if order.ID == mux.Vars(r)["id"] {
				err = c.WriteJSON(order)
				if err != nil {
					log.Println("err:", err)
					break
				}
			}
		}
	}
}
