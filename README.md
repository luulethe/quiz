# quiz logic server

## Run in local:

  go build -o app .

  DEPLOY=dev APP_LOG_PATH=logs ./app

## Build Docker and Run:

  docker build -t quiz-server:latest --build-arg BUILD_FOLDER="." .

  docker run --name quiz-server --network host --env DEPLOY=dev --env APP_LOG_PATH=/src/logs quiz-server:latest



# kafka-consumer

## Run in local:

  go build -o app ./kafka_service

  DEPLOY=dev APP_LOG_PATH=logs ./app

## Build Docker and Run:

  docker build -t quiz-kafka-consumer:latest --build-arg BUILD_FOLDER="kafka_service" .

  docker run --name quiz-kafka-consumer --network host --env DEPLOY=dev --env APP_LOG_PATH=/src/logs quiz-kafka-consumer:latest

