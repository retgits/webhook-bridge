package pubnub

import (
	nats "github.com/nats-io/go-nats"
	pubnub "github.com/pubnub/go"
)

// PubNub represents the configuration of how the app connects to PubNub and which channels and
// listeners are associated with it.
type PubNub struct {
	PubNubPublishKey   string            `required:"true"`
	PubNubSubscribeKey string            `required:"true"`
	PubNubChannels     []string          `required:"true"`
	PubNubNatsMap      map[string]string `required:"true"`
	Config             pubnub.Config
	Listeners          []*pubnub.Listener
	PubNubServer       *pubnub.PubNub
	NatsConn           *nats.Conn
}
