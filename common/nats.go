package common

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	nats "github.com/nats-io/go-nats"
)

const (
	// NatsDeadLetterQueue is the default NATS queue for messages that cannot be delivered
	NatsDeadLetterQueue = "dlq"
)

type natsConfig struct {
	NatsName string `required:"true"`
	NatsURL  string `required:"true"`
}

// NatsConnect handles connecting the a NATS server. It returns a
// NATS connection based on the environment variables:
// - NATSNAME: The name of the NATS connection
// - NATSURL: The URL of the NATS server to connect to
func NatsConnect() (*nats.Conn, error) {
	var c natsConfig
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}

	opts := []nats.Option{nats.Name(c.NatsName)}

	nc, err := nats.Connect(c.NatsURL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	return nc, nil
}

// NatsStop handles gracefully shutting down any connections to the NATS server
func NatsStop(nc *nats.Conn) {
	nc.Close()
}
