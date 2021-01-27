FROM golang:1.14

WORKDIR /go/src/webCrawler
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

