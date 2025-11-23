package http

import (
	"github.com/ThreeDotsLabs/watermill/message"
)

type Handler struct {
	pub message.Publisher
}
