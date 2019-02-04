# Webhook Bridge

![webhook-bridge](./icons/webhook-bridge-256.png)

Webhooks are an awesome way to get notified when something happens, like when a new Pull Request is created in a GitHub repo or when a new commit is done. The problem is when your build server is running on a machine that doesn't have a public IP address.

Webhook Bridge is a collection of small agents, connected through [NATS](http://nats.io), that connect the outside world securely to your inside network (_no open firewall ports_).

## Available agents

* PubNub: An ingress agent that connects to [PubNub](https://www.pubnub.com/) and leverages the awesome functionality they provide. Using the [REST API](https://www.pubnub.com/docs/pubnub-rest-api-documentation) to publish to PubNub and the [Go SDK](https://www.pubnub.com/docs/go/pubnub-go-sdk) to receive messages and publish them to NATS.
* Jenkins: An egress agent that connects to a NATS topic and triggers builds in Jenkins

## Getting the sources

```bash
git clone https://github.com/retgits/webhook-bridge
```

## Building the agents

The [Makefile](./Makefile) relies on Go Modules from [GoCenter](https://gocenter.io) to build the agents. To make using GoCenter as easy as possible, the Makefile relies on `goc` which you can install using `make setup-tools` or through the [goc](https://github.com/jfrog/goc) repo.

| Makefile target        | Description                                           |
|------------------------|-------------------------------------------------------|
| compile-jenkins-docker | Compiles the Jenkins agent and builds a Docker image. |
| compile-jenkins-mac    | Compiles the Jenkins agent to run on macOS.           |
| compile-pubnub-docker  | Compiles the PubNub agent and builds a Docker image.  |
| compile-pubnub-mac     | Compiles the PubNub agent to run on macOS.            |

## Other Makefile targets

The Makefile has a few other useful targets as well

| Makefile target  | Description                                                                                                |
|------------------|------------------------------------------------------------------------------------------------------------|
| fmt              | Fmt runs the commands 'gofmt -l -w' and 'gofmt -s -w' and prints the names of the files that are modified. |
| go-score         | Get a score based on GoReportcard.                                                                         |
| go-test-coverage | Run all test cases and generate a coverage report.                                                         |
| go-test          | Run all testcases.                                                                                         |
| lint             | Lint examines Go source code and prints style mistakes for all packages.                                   |
| setup-deps       | Get all the Go dependencies.                                                                               |
| setup-test       | Make preparations to be able to run tests.                                                                 |
| setup-tools      | Get the tools needed to test and validate the project.                                                     |
| test             | Run all test targets.                                                                                      |
| vet              | Vet examines Go source code and reports suspicious constructs.                                             |

## Running the app

### PubNub Agent

The PubNub Agent needs a few environment variables to be able to work

| Variable           | Description                                                                                               |
|--------------------|-----------------------------------------------------------------------------------------------------------|
| LOGLEVEL           | The [zerolog](https://github.com/rs/zerolog) loglevel to use (like `debug` or `info`)                     |
| PUBNUBPUBLISHKEY   | The publisher key for PubNub                                                                              |
| PUBNUBSUBSCRIBEKEY | The subscriber key for PubNub                                                                             |
| PUBNUBCHANNELS     | A comma separated list of PubNub channels to listen to                                                    |
| PUBNUBNATSMAP      | A comma separated list of key value pairs to match PubNub channels to NATS topics (like `github:jenkins`) |
| NATSNAME           | The name of your subscriber to connect to NATS                                                            |
| NATSURL            | The URL to connect to NATS (like `nats://localhost:4222`)                                                 |

To get a PubNub account, simply go to their [dashboard](https://dashboard.pubnub.com/login) and use “_SIGN UP_” to create a new account. After signing up, use the big red button to create a new app (the name doesn’t matter, and if you want you can change it later too). Now click on the newly created app and you’ll see a new KeySet. The Publish and Subscriber key are quite important as those are the ones you need for the configuration.

### Jenkins Agent

The Jenkins Agent needs a few environment variables to be able to work

| Variable           | Description                                                                                               |
|--------------------|-----------------------------------------------------------------------------------------------------------|
| LOGLEVEL           | The [zerolog](https://github.com/rs/zerolog) loglevel to use (like `debug` or `info`)                     |
| JENKINSUSERNAME    | The username to connect to Jenkins                                                                        |
| JENKINSAPITOKEN    | The API token to connect to Jenkins                                                                       |
| JENKINSBASEURL     | The URL to connect to Jenkins (like `http://myjenkinsserver:8080/job/%s/build`)                           |
| NATSNAME           | The name of your subscriber to connect to NATS                                                            |
| NATSURL            | The URL to connect to NATS (like `nats://localhost:4222`)                                                 |

To get the _apitoken_ in Jenkins

1. Log in to Jenkins.
2. Click you name (upper-right corner).
3. Click Configure (left-side menu).
4. Use "Add new Token" button to generate a new one then name it.

## License

See the [LICENSE](./LICENSE) file in the repository

## Icon

The amazing icon is made by [Freepik](https://www.freepik.com/) from [www.flaticon.com](https://www.flaticon.com/) and is licensed by [CC 3.0 BY](http://creativecommons.org/licenses/by/3.0/)