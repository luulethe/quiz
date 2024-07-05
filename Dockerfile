FROM golang:1.20 AS builder

RUN apt-get update \
 && apt-get install -y ca-certificates \
 && apt-get install -y protobuf-compiler=3.6.* \
 && GO111MODULE=on go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.25.0 google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.0.1

ARG BUILD_FOLDER

WORKDIR /src
ADD . /src
RUN make proto-gen \
    && cd $BUILD_FOLDER \
    && GO111MODULE=on go build -mod=vendor -o app .

FROM debian:buster-20191014-slim

RUN apt-get update && apt-get install -y ca-certificates procps curl

ENV APP_IN_K8S="true"
ARG BUILD_FOLDER
ADD config /config
ADD . /src
COPY --from=builder /src/$BUILD_FOLDER/app /bin/app
CMD /bin/app