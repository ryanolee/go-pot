# go-pot 🍯
A HTTP tarpit written in Go designed to maximize bot misery through very slowly feeding them an infinite stream of fake secrets. (Originally the subject of a talk)

<img src="docs/img/gopher.png" width="400px" />

## Features
- **Realistic output**: Go pot will respond to requests with an infinite stream of realistic looking, parseable structured data full of fake secrets. `xml`, `json`, `yaml`, `hcl`, `toml`, `csv`, `ini`, and `sql` are all supported.
- **Intelligent stalling**: Go pot will attempt to work out how long a bot is willing to wait for a response and stall for exactly that long. This is done gradually making requests slower and slower until a timeout is reached. (Or the bot hangs forever!)
- **Small Profile**: Go pot can run on extremely low resource machines and is designed to be as lightweight as possible.
- **Clustering Support**: Go pot can be run in a clustered mode where multiple instances can share information about how long bots are willing to wait for a response. Also in cluster mode nodes can be configured to restart / reallocate IP addresses to avoid being blacklisted by connecting clients.
- **Customizable**: Go pot can be customized to respond with different different response times support for more protocols is planned.

## Installation
Go pot is distributed as a standalone go binary or docker image. You can download the latest release from the [releases page](https://github.com/ryanolee/go-pot/releases). Docker images are available on the [ghcr.io registry](https://github.com/ryanolee/go-pot/pkgs/container/go-pot).

### Docker
In order to run an example instance of go-pot using docker, you can use the following command:
```bash
docker run -p 8080:8080 --rm ghcr.io/ryanolee/go-pot:latest --host=0.0.0.0 --port=8080
```
See the `examples` directory for more examples of how to run go-pot in various configurations.

### Standalone
In order to run go-pot as a standalone binary, you can download the latest release from the [releases page](https://github.com/ryanolee/go-pot/releases) and run it with the following command:
```bash
./go-pot start
```
Then visit `http://localhost:8080` in your browser to see the go-pot in action.

## Configuration
Configuration for go-pot follows the following order of precedence (From lowest to highest):
 * **Defaults**: Default values can be found in the [config/default.go](config/default.go) file.
 * **Config file**: A configuration file path can be specified using the `--config-file` flag or using the `GO_POT__CONFIG_FILE` environment variable. An example reference configuration file can be found in the [examples/config/reference.yml](examples/config/reference.yml) file.
 * **Command line flags**: Command line flags can be used to override configuration values. Run `go-pot --help` to see a list of available flags.
 * **Environment variables**: Environment variables can be used to override configuration values. Environment variables are prefixed with `GO_POT__` and deliminated with "__"'s for further keys. For instance `server.host` can be overridden with `GO_POT__SERVER__HOST`. 

## Deployment
Go pot can be deployed in a variety of ways. See the [cdk](cdk) directory for an example of how to deploy go-pot using the AWS CDK on ECS Fargate for which it has native clustering support.

## Contributing
Contributions are welcome! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) file for more information on how to contribute.

See the internal [INTERNALS.md](INTERNALS.md) file for more information on how go-pot works.

## Credits
Go pot was originally inspired by the (Reverse slow loris)[https://github.com/nickhuber/reverse-slowloris] project by [Nick Huber](https://github.com/nickhuber)
The go pot logo created by `@_iroshi` and is licensed under the [CC0](https://creativecommons.org/publicdomain/zero/1.0/) license.
