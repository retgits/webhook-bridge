# Webhook Bridge

![webhook-bridge](./icons/webhook-bridge-256.png)

Webhooks are an awesome way to get notified when something happens, like when a new Pull Request is created in a GitHub repo or when a new commit is done. The problem is when your build server is running on a machine that doesn't have a public IP address.

**Webhook Bridge** is a small server that connects to [PubNub](https://www.pubnub.com/) and leverages the awesome functionality they provide. **Webhook Bridge** uses the [REST API](https://www.pubnub.com/docs/pubnub-rest-api-documentation) to publish to PubNub and the [Go SDK](https://www.pubnub.com/docs/go/pubnub-go-sdk) to receive messages.

## Getting the sources

```bash
go get -u github.com/retgits/webhook-bridge
```

## Building the app

```bash
make build
```

## Running the app

### Configuration

The app looks for a `config.yml` file in three places:

* `/etc/webhookbridge`
* `$HOME/.webhookbridge`
* `<working directory>`

The configuration, of which `config_template.yml` is a good template, contains of several key/value pairs

```yml
loglevel:      ## The loglevel to use. Allowed values are: debug, info, warn, error, fatal, panic
jenkins:
  username:    ## The username to connect to Jenkins
  apitoken:    ## The API key to connect to Jenkins
  baseurl:     ## The full URL for Jenkins, with %s as the prameter for the job name
pubnub:
  keys:
    publish:   ## The key to publish to PubNub
    subscribe: ## The key to subscribe to PubNub
  channels:    ## An array of channels to receive messages on
    - xyz
```

Additionally, thanks to [viper](https://github.com/spf13/viper), you can also provide the configuration variables through env vars

```bash
export LOGLEVEL=debug
export JENKINS_USERNAME=<username>
export JENKINS_APITOKEN=<apitoken>
export JENKINS_BASEURL=<url>
export PUBNUB_KEYS_PUBLISH=<key>
export PUBNUB_KEYS_SUBSCRIBE=<key>
export PUBNUB_CHANNELS=<chan1>,<chan2>
```

### Jenkins

To get the _apitoken_ in Jenkins

1. Log in to Jenkins.
1. Click you name (upper-right corner).
1. Click Configure (left-side menu).
1. Use "Add new Token" button to generate a new one then name it.

The _baseurl_ follows the structure of `"http://myjenkinsserver:8080/job/%s/build"`

### PubNub

To get a PubNub account, simply go to their [dashboard](https://dashboard.pubnub.com/login) and use “_SIGN UP_” to create a new account. After signing up, use the big red button to create a new app (the name doesn’t matter, and if you want you can change it later too). Now click on the newly created app and you’ll see a new KeySet. The Publish and Subscriber key are quite important as those are the ones you need to set in the `config.yml`

### Start

Simply run `./webhook-bridge`!

## License

See the [LICENSE](./LICENSE) file in the repository

## Icon

The amazing icon is made by [Freepik](https://www.freepik.com/) from [www.flaticon.com](https://www.flaticon.com/) and is licensed by [CC 3.0 BY](http://creativecommons.org/licenses/by/3.0/)