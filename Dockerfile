FROM golang:1.14

RUN apt-get update

RUN apt-get install -y apt-file

RUN apt-file update

RUN apt-get install -y vim

WORKDIR /go/src/app
COPY . .

RUN go build

CMD go run main.go