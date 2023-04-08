FROM golang:latest AS build

# Set the working directory to the root of the project
WORKDIR /app

# Install the protoc compiler
RUN apt-get update && apt-get install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

COPY ./src/go.mod ./src/go.sum ./
RUN go mod download

# Then copy the entire project to the container
COPY ./src .

# Build client-server RPCs
WORKDIR /app/pb
RUN protoc --experimental_allow_proto3_optional --go_out=./ --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_out=./ comm.proto

# Build server-server RPCs
WORKDIR /app/server/pb/
RUN protoc --experimental_allow_proto3_optional --go_out=./ --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_out=./ --proto_path=./../../pb/ --proto_path=./ replication.proto

# Build the client and server binaries
WORKDIR /app
RUN go build -o ./bin/client ./client 
RUN go build -o ./bin/server ./server 

# Client image to copy the client binary to the host
FROM ubuntu:latest AS client

WORKDIR /app
COPY --from=build /app/bin/client client
ENTRYPOINT ["/app/client"]


# Server image to copy the server binary to the host
FROM ubuntu:latest AS server

WORKDIR /app
COPY --from=build /app/bin/server server
ENTRYPOINT ["/app/server"]


FROM postgres:latest AS db

ENV POSTGRES_db=postgres
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres

WORKDIR /docker-entrypoint-initdb.d
COPY ./src/server/db/init.sql ./