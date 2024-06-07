# Go Pot Examples: Metrics
This example covers putting a reverse proxy in front of a go-pot instance.

## Running the Example
To run the example you will need docker and docker-compose installed on your machine.
To start the example, run the following commands **in the project root** :
```bash
docker compose -f examples/reverse-proxy/docker-compose-reverse-proxy-example.yml up
```

Visit localhost:8080 to see go-pot running behind a reverse proxy. (Nginx in this case)

You can then visit localhost:8081 to visit go-pot directly.

to stop the example, run the following command **in the project root** :
```bash
docker compose -f examples/reverse-proxy/docker-compose-reverse-proxy-example.yml down
```
