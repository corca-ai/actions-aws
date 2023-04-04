FROM golang:1.19-alpine as golang
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o server cmd/actions-ec2.go

FROM alpine:3.14
WORKDIR /app

COPY --from=golang /app/server /app/server

CMD [ "/app/server" ]
