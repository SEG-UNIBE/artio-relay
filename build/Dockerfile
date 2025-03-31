FROM golang:latest

WORKDIR /app
COPY . /app

RUN go mod tidy
RUN go mod download
RUN go build -o /app/bin/relay /app/cmd/main.go

EXPOSE 8000

CMD ["/app/bin/relay"]