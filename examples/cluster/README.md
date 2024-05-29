# Go Pot Examples: Cluster
This example demonstrates how to use the "cluster" functionality of Go Pot. Clustering allows each instance of Go Pot to warn other instances about timeouts for certain requesting IP addresses. 

## Running the Example
To run the example you will need docker and docker-compose installed on your machine.
To start the example, run the following commands **in the project root** :
```bash
docker compose -f examples/cluster/docker-compose-cluster-example.yml up
```

This will start three instances of Go Pot binding to ports 8080, 8081, and 8082. Each instance will share information about "known" timeouts from the other instances.