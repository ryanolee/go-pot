# Go Pot Examples: Ftp
This example covers running go-pot as an FTP Server.

## Running the Example
To run the example you will need docker and docker-compose installed on your machine.
To start the example, run the following commands **in the project root** :
```bash
docker compose -f examples/ftp/docker-compose-ftp.yml up
```

Visit localhost:2121 to with an FTP client of your choice (FileZilla has been tested) to see, Note that any password and username combination will work.


to stop the example, run the following command **in the project root** :
```bash
docker compose -f examples/ftp/docker-compose-ftp.yml down
```
