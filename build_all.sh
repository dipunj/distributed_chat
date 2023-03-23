# /bin/sh

docker build -t chat-server -f Dockerfile --target server .
docker build -t chat-client -f Dockerfile --target client .
docker build -t chat-db -f Dockerfile --target db .