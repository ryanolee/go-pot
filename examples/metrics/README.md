# Go Pot Examples: Metrics
This example covers the prometheus metrics functionality of Go Pot.

## Running the Example
To run the example you will need docker and docker-compose installed on your machine.
To start the example, run the following commands **in the project root** :
```bash
docker compose -f examples/metrics/docker/docker-compose-metrics-example.yml up
```

Visit localhost:8080 and localhost:8081 to trigger the running go-pot instances to generate metrics.

You can then visit localhost:9090 to see the prometheus dashboard. (`time_wasted` and `secrets_generated` can be viewed in the metrics tab)

to stop the example, run the following command **in the project root** :
```bash
docker compose -f examples/metrics/docker/docker-compose-metrics-example.yml down
```
