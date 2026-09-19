FROM golang:1.20-alpine

WORKDIR WORKDIR /go/src/app
COPY . .
RUN go build -o /bin/lococ-job ./cmd/lococ-job
RUN go build -o /bin/lococ-web ./cmd/lococ-web

CMD ["/bin/lococ-web"]
