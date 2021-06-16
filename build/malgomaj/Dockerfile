# syntax=docker/dockerfile:1
FROM golang:1.16 AS builder
WORKDIR /go/src/github.com/self-host/self-host/
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o malgomaj github.com/self-host/self-host/cmd/malgomaj

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /usr/bin
COPY --from=builder /go/src/github.com/self-host/self-host/malgomaj .
CMD ["/usr/bin/malgomaj"]
