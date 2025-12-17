package main

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type FollowRequestSent struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type EventsCounter interface {
	CountEvent() error
}

type EventHandler struct {
	counter EventsCounter
}

func (h *EventHandler) Handle(ctx context.Context, event *FollowRequestSent) error {
	return h.counter.CountEvent()
}

func NewFollowRequestSentHandler(counter EventsCounter) cqrs.EventHandler {
	handler := EventHandler{counter: counter}
	return cqrs.NewEventHandler(
		"FollowRequestSentHandler",
		handler.Handle,
	)
}
