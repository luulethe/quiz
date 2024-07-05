# grpc_sample

## Run in local:

  GO111MODULE=on go build -mod=vendor -o app .

  DEPLOY=dev APP_LOG_PATH=logs ./app

## Build Docker and Run:

  docker build -t go-grpc-sample:latest --build-arg BUILD_FOLDER="." .

  docker run --name go-grpc-sample --network host --env DEPLOY=dev --env APP_LOG_PATH=/src/logs go-grpc-sample:latest


# Cron

## Run in local:

  GO111MODULE=on go build -mod=vendor -o app ./job_service

  DEPLOY=dev APP_LOG_PATH=logs ./app

## Build Docker and Run:

  docker build -t go-cron-sample:latest --build-arg BUILD_FOLDER="job_service" .

  docker run --name go-cron-sample --network host --env DEPLOY=dev --env APP_LOG_PATH=/src/logs go-cron-sample:latest


# kafka-consumer

## Run in local:

  GO111MODULE=on go build -mod=vendor -o app ./kafka_service

  DEPLOY=dev APP_LOG_PATH=logs ./app

## Build Docker and Run:

  docker build -t go-kafka-consumer-sample:latest --build-arg BUILD_FOLDER="kafka_service" .

  docker run --name go-kafka-consumer-sample --network host --env DEPLOY=dev --env APP_LOG_PATH=/src/logs go-kafka-consumer-sample:latest

