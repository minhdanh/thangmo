FROM golang:1.16-alpine

WORKDIR WORKDIR /go/src/app
COPY . .
RUN go build -o /bin/thangmo-job ./cmd/thangmo-job
RUN go build -o /bin/thangmo-web ./cmd/thangmo-web

CMD ["/bin/thangmo-web"]
