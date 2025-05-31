#!/bin/sh

ROOT=$(git rev-parse --show-toplevel)
cd $ROOT || exit

# build
go build .
go build ./cmd/sink
go build ./cmd/source

# clean
kill -9 $(lsof -t -i:8080)
kill -9 $(lsof -t -i:8000)
kill -9 $(lsof -t -i:8001)
kill -9 $(lsof -t -i:8002)
kill -9 $(lsof -t -i:8003)

# run services
./webhook-dispatcher -port 8080 &
./sink -port 8000 &
./sink -port 8001 &
./sink -port 8002 &
./sink -port 8003 &

sleep 2

# setup webhooks-dispatcher
curl -X POST "http://localhost:8080/subscribers" --data '{"url":"http://localhost:8000", "token":"token", "matching_rules": {"url_path": "even"}}'
curl -X POST "http://localhost:8080/subscribers" --data '{"url":"http://localhost:8001", "token":"token", "matching_rules": {"url_path": "odd"}}'
curl -X POST "http://localhost:8080/subscribers" --data '{"url":"http://localhost:8002", "token":"token", "matching_rules": {"url_path": "even"}}'
curl -X POST "http://localhost:8080/subscribers" --data '{"url":"http://localhost:8003", "token":"token", "matching_rules": {"url_path": "odd"}}'

#test
./source -dispatcherURL http://localhost:8080 &
PID=$!
sleep 10
echo "Killing dispatcher"
kill $PID

# shutdown
kill -9 $(lsof -t -i:8080)
kill -9 $(lsof -t -i:8000)
kill -9 $(lsof -t -i:8001)
kill -9 $(lsof -t -i:8002)
kill -9 $(lsof -t -i:8003)
