FROM golang:1.21

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o /TZ-ONE

EXPOSE 9091

CMD ["TZ-ONE"]