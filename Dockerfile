FROM golang:1.19 as golang
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o server cmd/actions-ec2.go

CMD [ "/app/server" ]
