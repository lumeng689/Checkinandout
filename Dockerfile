FROM golang:1.14

WORKDIR /go/src/app
COPY . .

RUN go build

CMD go run main.go