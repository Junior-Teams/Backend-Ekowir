FROM golang:1.23-alpine

WORKDIR /usr/src/app

RUN go install github.com/air-verse/air@v1.52.3

COPY . ./
RUN go mod tidy

CMD [ "air", "-c", ".air.toml" ]



