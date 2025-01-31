# Go - Pot Internals
Go pot as the names suggests is surprisingly 🥁... written in go! It is a HTTP tarpit made up of a few different components.

## Dependencies
* **Fiber** (http router)[https://docs.gofiber.io/]
* **Fx** (dependency injection)[https://uber-go.github.io/fx/]
* **Zap** (logging)[https://pkg.go.dev/go.uber.org/zap]
* **Cobra** (CLI)[https://pkg.go.dev/github.com/spf13/cobra]
* **koanf** (configuration)[https://github.com/knadh/koanf]
* **go-chaff** (fake structured data generation)[https://github.com/ryanolee/go-chaff]
* **go-faker** (fake data generation)[https://github.com/go-faker/faker]
* **memberlist** (inter-node communication)[https://github.com/hashicorp/memberlist]

## Components
Go pot is made up of a few different components that come together to make the staller. Some of the more idiomatic components are:
* **Staller**: A special handler that will stall for a request for a given amount of time. It gets a generator instance it will keep on calling for new data until just before the timeout it has been given is reached. At which point it will correctly terminate the response.
* **Generator**: A generator will provide an infinite stream of fake structured data. That can be serialized into a number of different formats.
* **TimeoutWatcher**: The timeout watcher will keep track of how long a bot is willing to wait for a response. It will do this by watching when a given IP address disconnects. If it gets a few similar disconnects in a row it will assume that that is the maximum time a bot is willing to wait for a response and then give a time just under that to the staller.
* **Cluster**: The cluster is a way of sharing information about how long bots are willing to wait for a response to other nodes in the cluster. It uses memberlist (go)
* **Recast**: Recast is a way of restarting / reallocating IP addresses to avoid being blacklisted by connecting clients. It uses telemetry to see if stalling connections and moves to a different IP block if not. 
* **Detect / Multi protocol listener** Detect aims to watch for traffic on a TCP listener and make a guess at which protocol data being sent down the pipe belongs to. It does this by.
When a new connection is opened:
 * Wait for some data to be sent by the client
 * If no data is sent in X seconds begin to "probe" by sending different protocol headers back down the pipe
 * Wait for data while probes are sent 
 * If still no data is sent change to the fallback handler
