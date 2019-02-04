package jenkins

import nats "github.com/nats-io/go-nats"

// Jenkins represents the configuration of how to connect to a Jenkins server using username and
// API token for authentication. The BaseURL follows the structure of `"http://myjenkinsserver:8080/job/%s/build"`
type Jenkins struct {
	JenkinsUsername string `required:"true"`
	JenkinsAPIToken string `required:"true"`
	JenkinsBaseURL  string `required:"true"`
	NatsChannel     string `required:"true"`
	NatsConn        *nats.Conn
}
