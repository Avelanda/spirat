FROM golang:1.20.3-bullseye

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["spirat"]
