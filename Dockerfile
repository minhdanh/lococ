FROM golang:1.20-alpine

WORKDIR /go/src/app
COPY . .
RUN go build -o /bin/lococ ./cmd/lococ

CMD ["/bin/lococ"]
