# go-pot 🍯
A HTTP tarpit written in Go designed to maximize bot misery through very slowly feeding them an infinite stream of fake secrets.

<img src="docs/img/gopher.png" width="400px" />

## Features
- **Realistic output**: Go pot will respond to requests with an infinite stream of realistic looking, parseable structured data full of fake secrets. `xml`, `json`, `yaml`, `hcl`, `toml`, `csv`, `ini`, and `sql` are all supported.
- **Intelligent stalling**: Go pot will attempt to work out how long a bot will wait for a response and stall for exactly that long.

## Installation
Go pot is distributed as a standalone go binary or docker image. You can download the latest release from the [releases page](https://github.com/ryanolee/go-pot/releases). Docker images are available on the [ghcr.io registry](https://github.com/ryanolee/go-pot/pkgs/container/go-pot).

### Docker
In order to run an example instance of go-pot using docker, you can use the following command:
```bash
docker run -p 8080:8080 --rm ghcr.io/ryanolee/go-pot:latest --host=0.0.0.0 --port=8080
```

See the `examples` directory for more examples of how to run go-pot in various configurations.

## Configuration
Configuration for go-pot follows the following order of precedence (From lowest to highest):
 * **Defaults**: Default values can be found in the [config/default.go](config/default.go) file.
 * **Config file**: A configuration file path can be specified using the `--config-file` flag or using the `GO_POT__CONFIG_FILE` environment variable. An example reference configuration file can be found in the [examples/config/reference.yml](examples/config/reference.yml) file.
 * **Command line flags**: Command line flags can be used to override configuration values. Run `go-pot --help` to see a list of available flags.
 * **Environment variables**: Environment variables can be used to override configuration values. Environment variables are prefixed with `GO_POT__` and deliminated with "__"'s for further keys. For instance `server.host` can be overridden with `GO_POT__SERVER__HOST`. 

