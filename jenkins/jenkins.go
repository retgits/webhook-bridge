package jenkins

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
	nats "github.com/nats-io/go-nats"
	"github.com/retgits/webhook-bridge/common"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

// Register is the function that creates a new instance of the Jenkins agent.
func Register() (*Jenkins, error) {
	var jenkins Jenkins
	err := envconfig.Process("", &jenkins)
	if err != nil {
		return nil, err
	}

	conn, err := common.NatsConnect()
	if err != nil {
		return nil, err
	}

	jenkins.NatsConn = conn

	return &jenkins, nil
}

// Start is the function that takes care of starting the agent and makes sure a file is creating to perform healthchecks
// in case the agent runs in a Docker container. By returning a boolean channel, it will make sure to create an infinite
// loop that can be ended grafecully by calling the stop method.
func (j *Jenkins) Start() chan bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		runfile, err := os.Create(".running")
		if err != nil {
			log.Error().Msgf("Error while creating .runningfile: %s\nthis means the container is unhealthy but the server will work...", err.Error())
		}
		log.Info().Msgf("Created .running file: %v", runfile)
		runfile.Close()
	}
	// Start an indefinite loop to receive messages
	done := make(chan bool)
	j.NatsConn.Subscribe(j.NatsChannel, func(msg *nats.Msg) {
		j.handleMessage(msg)
	})
	return done
}

// Stop takes care of the actions that need to be performed when gracefully shutting down the agent
func (j *Jenkins) Stop() {
	log.Info().Msg("Closing connection to NATS")
	common.NatsStop(j.NatsConn)
}

// handleMessage handles all regular messages that are coming through.
func (j *Jenkins) handleMessage(msg *nats.Msg) {
	jsonstring := string(msg.Data)
	ref := gjson.Get(jsonstring, "ref")
	repo := gjson.Get(jsonstring, "repository.name")

	if len(ref.String()) > 0 && len(repo.String()) > 0 && ref.String() == "master" {
		log.Info().Msgf("Received Push event for: %s", repo.String())
		j.triggerJob(repo.String())
	} else {
		log.Info().Msgf("Handling an invalid message:\n%s", jsonstring)
	}
}

// triggerJob triggers a Jenkins job based on the name of the job configured in Jenkins
func (j *Jenkins) triggerJob(job string) error {
	authString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", j.JenkinsUsername, j.JenkinsAPIToken)))
	url := fmt.Sprintf(j.JenkinsBaseURL, job)

	log.Debug().Msgf("Sending request to: %s", url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("authorization", fmt.Sprintf("Basic %s", authString))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Received data from Jenkins:\n%+v\n%s", res, string(body))
	return nil
}
