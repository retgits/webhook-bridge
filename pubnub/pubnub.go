package pubnub

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	pubnub "github.com/pubnub/go"
	"github.com/pubnub/go/utils"
	"github.com/retgits/webhook-bridge/common"
	"github.com/rs/zerolog/log"
)

// Register is the function that creates a new instance of the PubNub agent.
func Register() (*PubNub, error) {
	var pn PubNub
	err := envconfig.Process("", &pn)
	if err != nil {
		return nil, err
	}

	// Create a PubNub struct
	pn.Config = pubnub.Config{
		Origin:                     "ps.pndsn.com",
		Secure:                     true,
		UUID:                       fmt.Sprintf("pn-%s", utils.UUID()),
		ConnectTimeout:             10,
		NonSubscribeRequestTimeout: 10,
		SubscribeRequestTimeout:    310,
		MaximumLatencyDataAge:      60,
		MaximumReconnectionRetries: 50,
		SuppressLeaveEvents:        false,
		DisablePNOtherProcessing:   false,
		PNReconnectionPolicy:       pubnub.PNNonePolicy,
		MessageQueueOverflowCount:  100,
		MaxIdleConnsPerHost:        30,
		MaxWorkers:                 20,
		PublishKey:                 pn.PubNubPublishKey,
		SubscribeKey:               pn.PubNubSubscribeKey,
	}

	conn, err := common.NatsConnect()
	if err != nil {
		return nil, err
	}

	pn.NatsConn = conn

	return &pn, nil
}

// Start is the function that takes care of starting the agent and makes sure a file is creating to perform healthchecks
// in case the agent runs in a Docker container. By returning a boolean channel, it will make sure to create an infinite
// loop that can be ended grafecully by calling the stop method.
func (pn *PubNub) Start() chan bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		runfile, err := os.Create(".running")
		if err != nil {
			log.Error().Msgf("Error while creating .runningfile: %s\nthis means the container is unhealthy but the server will work...", err.Error())
		}
		log.Info().Msgf("Created .running file: %v", runfile)
		runfile.Close()
	}
	log.Info().Msg("Connecting to PubNub")
	client := pubnub.NewPubNub(&pn.Config)
	pn.PubNubServer = client

	// Register listeners
	log.Info().Msg("Registering Listeners")
	listener := pubnub.NewListener()
	client.AddListener(listener)
	listeners := make([]*pubnub.Listener, 1)
	listeners[0] = listener

	// Subscribe to channels
	log.Info().Msg("Subscribing to Channels Listeners")
	client.Subscribe().Channels(pn.PubNubChannels).Execute()

	// Start an indefinite loop to receive messages
	done := make(chan bool)
	go func() {
		for {
			select {
			case status := <-listener.Status:
				pn.handleStatusMessage(status)
			case message := <-listener.Message:
				pn.handleMessage(message)
			case <-listener.Presence:
				// TODO: Precense allows you to subscribe to realtime Presence events, such as join, leave, and timeout, by UUID. This is currently not implemented
			case <-done:
				// listen for the exit signal and break
				break
			}
		}
	}()
	return done
}

// Stop takes care of the actions that need to be performed when gracefully shutting down the agent
func (pn *PubNub) Stop() {
	log.Info().Msg("Removing all PubNub Listeners")
	for _, listener := range pn.Listeners {
		pn.PubNubServer.RemoveListener(listener)
	}
	log.Info().Msg("Unsubscribing from all PubNub Channels")
	pn.PubNubServer.UnsubscribeAll()
	log.Info().Msg("Closing connection to NATS")
	common.NatsStop(pn.NatsConn)
}

// handleStatusMessage handles all the status messages that are sent through PubNub
func (pn *PubNub) handleStatusMessage(status *pubnub.PNStatus) {
	switch status.Category {
	case pubnub.PNDisconnectedCategory:
		log.Info().Msg("Received status event: [pubnub.PNDisconnectedCategory], this is the expected category for an unsubscribe. This means there was no error in unsubscribing from everything")
	case pubnub.PNConnectedCategory:
		log.Info().Msg("Received status event: [pubnub.PNConnectedCategory], this is expected for a subscribe, this means there is no error or issue whatsoever")
	case pubnub.PNReconnectedCategory:
		log.Info().Msg("Received status event: [pubnub.PNReconnectedCategory], this usually occurs if subscribe temporarily fails but reconnects. This means there was an error but there is no longer any issue")
	case pubnub.PNAccessDeniedCategory:
		log.Info().Msg("Received status event: [pubnub.PNAccessDeniedCategory], this means that PAM does allow this client to subscribe to this channel and channel group configuration. This is another explicit error")
	}
}

// handleMessage handles all regular messages that are coming through.
func (pn *PubNub) handleMessage(message *pubnub.PNMessage) {
	jsonString, err := json.Marshal(message.Message)
	if err != nil {
		log.Error().Msgf("Error while marshalling PubNub message to JSON: %s", err.Error())
		return
	}

	var natsChannel string
	if channel, ok := pn.PubNubNatsMap[message.Channel]; !ok {
		log.Error().Msgf("Message received on channel [%s] has no matching NATS channel", message.Channel)
		natsChannel = common.NatsDeadLetterQueue
	} else {
		natsChannel = channel
	}

	log.Debug().Msgf("Received Message event:\n%v\nRequest was on channel: %s\nSending to NATS topic: %s", string(jsonString), message.Channel, natsChannel)
	err = pn.NatsConn.Publish(natsChannel, jsonString)
	if err != nil {
		log.Error().Msgf("Error while publishing message to NATS: %s", err.Error())
	}
}
