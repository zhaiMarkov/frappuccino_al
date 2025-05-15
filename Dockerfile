FROM golang:1.23

RUN apt-get update && apt-get install -y wait-for-it

WORKDIR /

COPY . .

RUN go build -o main ./cmd

EXPOSE 8080

CMD ["wait-for-it", "db:${DB_PORT}", "--","./main"]