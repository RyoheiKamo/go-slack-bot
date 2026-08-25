FROM golang:1.27

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

CMD ["go", "run", "./cmd/server"]
