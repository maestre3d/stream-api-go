package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/maestre3d/stream-api/internal/aggregate"
	"github.com/maestre3d/stream-api/internal/event"
	"github.com/maestre3d/stream-api/internal/persistence"
	"github.com/neutrinocorp/gluon"
)

type ResponseMessage struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func respondError(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(ResponseMessage{
		Message: err.Error(),
		Status:  http.StatusInternalServerError,
	})
}

type Http interface {
	MapRoutes(r *mux.Router)
}

type OrderHttp struct {
	repo persistence.Order
	bus  *gluon.Bus
}

const baseTransactionFee = .25

var _ Http = &OrderHttp{}

func NewOrderHttp(r persistence.Order, b *gluon.Bus) *OrderHttp {
	return &OrderHttp{
		repo: r,
		bus:  b,
	}
}

func (o *OrderHttp) MapRoutes(r *mux.Router) {
	r.Path("/users/{user_id}/orders").Methods(http.MethodPost).HandlerFunc(o.issue)
	r.Path("/orders/{id}").Methods(http.MethodGet).HandlerFunc(o.get)
	r.Path("/orders/{id}").Methods(http.MethodPut, http.MethodPatch).HandlerFunc(o.update)
}

func (o *OrderHttp) issue(w http.ResponseWriter, r *http.Request) {
	totalItems, _ := strconv.Atoi(r.PostFormValue("total_items"))
	grossTotal, _ := strconv.ParseFloat(r.PostFormValue("subtotal"), 64)
	fee := grossTotal * baseTransactionFee

	order := aggregate.Order{
		ID:             uuid.NewString(),
		UserID:         mux.Vars(r)["user_id"],
		TotalItems:     totalItems,
		TransactionFee: fee,
		NetTotal:       grossTotal + fee,
		Status:         aggregate.OrderStatusPreparing,
		LastUpdate:     time.Now().UTC(),
	}

	ev := event.OrderIssued{
		ID:             order.ID,
		UserID:         order.UserID,
		TotalItems:     order.TotalItems,
		TransactionFee: order.TransactionFee,
		NetTotal:       order.NetTotal,
		Status:         order.Status,
		IssuedAt:       order.LastUpdate,
	}

	if err := o.repo.Save(r.Context(), order); err != nil {
		respondError(w, err)
		return
	}

	if err := o.bus.PublishWithSubject(r.Context(), ev, order.ID); err != nil {
		respondError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(order)
}

func (o *OrderHttp) get(w http.ResponseWriter, r *http.Request) {
	order, err := o.repo.Get(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		respondError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(order)
}

func (o *OrderHttp) update(w http.ResponseWriter, r *http.Request) {
	order, err := o.repo.Get(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		respondError(w, err)
		return
	}

	order.Status = strings.ToUpper(r.PostFormValue("order_status"))
	order.LastUpdate = time.Now().UTC()

	ev := event.OrderUpdated{
		ID:        order.ID,
		UserID:    order.UserID,
		Status:    order.Status,
		UpdatedAt: order.LastUpdate,
	}

	if err = o.repo.Save(r.Context(), order); err != nil {
		respondError(w, err)
		return
	}

	if err = o.bus.PublishWithSubject(r.Context(), ev, order.ID); err != nil {
		respondError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(order)
}
