package main

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/message"
)

type AlarmClient interface {
	StartAlarm() error
	StopAlarm() error
}

const smokeDetected = "1"

func ConsumeMessages(sub message.Subscriber, alarmClient AlarmClient) {
	messages, err := sub.Subscribe(context.Background(), "smoke_sensor")
	if err != nil {
		panic(err)
	}

	for msg := range messages {
		val := string(msg.Payload)
		if val == smokeDetected {
			if err = alarmClient.StartAlarm(); err != nil {
				fmt.Printf("failed to start alarm. retrying. %v \n", err)
				msg.Nack()
			}
		}
		msg.Ack()
	}
}
