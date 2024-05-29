# go-pot 🍯
A honeypot written in Go designed to maximize bot misery through very slowly feeding them an infinite stream of fake secrets.

## Features
- **Realistic output**: Go pot will respond to requests with an infinite stream of realistic looking, parseable structured data full of fake secrets. `xml`, `json`, `yaml`, ``
- **Intelligent stalling**: Go pot will attempt to work out how long a bot will wait for a response and stall for exactly that long.
