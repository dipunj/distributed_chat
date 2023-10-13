# Distributed Group Chat

Language Used: Go(version go1.19.5) with gRPC and protobuf 3

Invoke all the commands in the directory where docker-compose.yml file is located (i.e project root directory).

To build the containers run

```bash
docker-compose build
```

or alternatively you can run

```bash
source build_all.sh
```

The `docker-compose.yml` file creates a docker network called `cs2510`, on which 8 containers run:
- 1 server_i
- database_i

To start the containers run

```bash
docker-compose up -d
```

To stop the containers run

```bash
docker-compose down
```

So if you want to add a new client container which connects to the server, you can do so by:

```bash
# chat-client is the name of the client image.
docker run -it --net=chat chat-client
```

To check the logs of the server

```bash
# chat_server is the name of the server container
docker container logs chat_server 
```

By default the server runs on port 3000, which is exposed to the host machine
The client binary can also be invoked by the following command in the client container

```bash
./client -host <server_addr/ip> -port <server_port_number>
```
