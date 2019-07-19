package vaultransport

import (
	"context"
	"encoding/json"

	"github.com/go-kit/kit/log"
	amqptransport "github.com/go-kit/kit/transport/amqp"
	"github.com/streadway/amqp"
	"github.com/williamzion/vault/pkg/vaultendpoint"
)

func NewSubscriberHandler(endpoints vaultendpoint.Set, ch *amqp.Channel, log log.Logger) *amqptransport.Subscriber {
	options := []amqptransport.SubscriberOption{
		amqptransport.SubscriberResponsePublisher(amqptransport.NopResponsePublisher),
		amqptransport.SubscriberErrorEncoder(amqptransport.ReplyErrorEncoder),
		// amqptransport.SubscriberBefore(
		// 	amqptransport.SetPublishExchange(""),
		// 	amqptransport.SetPublishKey(""),
		// 	amqptransport.SetPublishDeliveryMode(amqp.Persistent),
		// 	amqptransport.SetContentType("application/json"),
		// 	amqptransport.SetContentEncoding("utf-8"),
		// 	// amqptransport.SetAckAfterEndpoint(false),
		// ),
	}
	sub := amqptransport.NewSubscriber(
		endpoints.HashEndpoint,
		decodeAMQPHashRequest,
		amqptransport.EncodeJSONResponse,
		options...,
	)

	sub.ServeDelivery(ch)()
}

func decodeAMQPHashRequest(_ context.Context, d *amqp.Delivery) (interface{}, error) {
	var req vaultendpoint.HashRequest
	err := json.Unmarshal(d.Body, &req)
	return req, err
}
