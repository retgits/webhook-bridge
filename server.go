package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	pubnub "github.com/pubnub/go"
	"github.com/pubnub/go/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/webhooks.v3/github"
)

type Jenkins struct {
	Username string
	APIToken string
	BaseURL  string
}

type PubNub struct {
	Config       pubnub.Config
	Channels     []string
	Listeners    []*pubnub.Listener
	PubNubServer *pubnub.PubNub
}

type Server struct {
	Jenkins Jenkins
	PubNub  PubNub
}

func NewServer() *Server {
	jenkins := Jenkins{
		Username: viper.GetString("jenkins.username"),
		APIToken: viper.GetString("jenkins.apitoken"),
		BaseURL:  viper.GetString("jenkins.baseurl"),
	}

	pubnub := PubNub{
		Config: pubnub.Config{
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
			PublishKey:                 viper.GetString("pubnub.keys.publish"),
			SubscribeKey:               viper.GetString("pubnub.keys.subscribe"),
		},
		Channels: viper.GetStringSlice("pubnub.channels"),
	}

	return &Server{
		Jenkins: jenkins,
		PubNub:  pubnub,
	}
}

func (s *Server) Start() {
	log.Info().Msg("Starting server")
	log.Info().Msg("Connecting to PubNub")
	pn := pubnub.NewPubNub(&s.PubNub.Config)
	s.PubNub.PubNubServer = pn

	log.Info().Msg("Registering Listeners")
	listener := pubnub.NewListener()
	pn.AddListener(listener)
	listeners := make([]*pubnub.Listener, 1)
	listeners[0] = listener

	log.Info().Msg("Subscribing to Channels Listeners")
	pn.Subscribe().Channels(s.PubNub.Channels).Execute()

	for {
		select {
		case status := <-listener.Status:
			handleStatusMessage(status)
		case message := <-listener.Message:
			handleMessage(message, &s.Jenkins)
		case <-listener.Presence:
			// TODO: Precense allows you to subscribe to realtime Presence events, such as join, leave, and timeout, by UUID. This is currently not implemented
		}
	}
}

func (s *Server) Stop() {
	log.Info().Msg("Prepating to shutdown server")
	log.Info().Msg("Removing all PubNub Listeners")
	for _, listener := range s.PubNub.Listeners {
		s.PubNub.PubNubServer.RemoveListener(listener)
	}
	log.Info().Msg("Unsubscribing from all PubNub Channels")
	s.PubNub.PubNubServer.UnsubscribeAll()
}

func handleStatusMessage(status *pubnub.PNStatus) {
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

func handleMessage(message *pubnub.PNMessage, jenkins *Jenkins) {
	jsonString, err := json.Marshal(message.Message)
	if err != nil {
		log.Error().Msgf("Error while marshalling PubNub message to JSON: %s", err.Error())
		return
	}

	log.Debug().Msgf("Received Message event:\n%v", string(jsonString))

	var pushPayload github.PushPayload
	var pullRequestPayload github.PullRequestPayload

	if err := json.Unmarshal(jsonString, &pushPayload); err == nil && len(pushPayload.Ref) > 0 {
		log.Info().Msgf("Received Push event for: %s", pushPayload.Repository.Name)
		if strings.Contains(pushPayload.Ref, "master") {
			handleGitHubPush(&pushPayload, jenkins)
		} else {
			log.Info().Msg("Push event was not for master branch, so ignoring event")
		}
	} else if err := json.Unmarshal(jsonString, &pullRequestPayload); err == nil && len(pullRequestPayload.Action) > 0 {
		log.Info().Msgf("Received Pull Request event for: %s", pullRequestPayload.Repository.Name)
		if strings.Contains(pullRequestPayload.Action, "opened") {
			handleGitHubPR(&pullRequestPayload, jenkins)
		} else {
			log.Info().Msg("Pull Request event was not for a new PR, so ignoring event")
		}
	} else {
		log.Info().Msgf("Received unknown event:\n%+v", message.Message)
	}
}

func handleGitHubPush(payload *github.PushPayload, jenkins *Jenkins) {
	authString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", jenkins.Username, jenkins.APIToken)))
	url := fmt.Sprintf(jenkins.BaseURL, payload.Repository.Name)

	log.Debug().Msgf("Sending request to: %s", url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Error().Msgf("Error while creating request for Jenkins: %s", err.Error())
		return
	}
	req.Header.Add("authorization", fmt.Sprintf("Basic %s", authString))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Msgf("Error while sending request to Jenkins: %s", err.Error())
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error().Msgf("Error while receiving response from Jenkins: %s", err.Error())
		return
	}
	log.Debug().Msgf("Received data from Jenkins:\n%+v\n%s", res, string(body))
}

func handleGitHubPR(payload *github.PullRequestPayload, jenkins *Jenkins) {
	log.Info().Msg("No actions defined for Pull Requests yet...")
}
