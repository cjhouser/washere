# https://docs.docker.com/language/golang/build-images/
FROM golang:1.20-alpine
WORKDIR /work
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o run
COPY static ./static
CMD ["/work/run"]