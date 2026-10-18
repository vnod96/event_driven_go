package main

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
)

func PublishInTx(
	message *message.Message,
	tx *sqlx.Tx,
) error {
	publisher, err := watermillSQL.NewPublisher(
		tx,
		watermillSQL.PublisherConfig{
			SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
		},
		watermill.NewSlogLogger(nil),
	)
	if err != nil {
		return err
	}

	err = publisher.Publish("ItemAddedToCart", message)
	
	return err
}
