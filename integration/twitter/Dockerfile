FROM golang:1.20.0

WORKDIR /go/src/app

COPY . .

ENV GO111MODULE on

RUN go build -o api/api api/main.go && \
	go build -o counter/counter counter/main.go && \
	go build -o votes/votes votes/main.go

CMD ["./api/api"]
