FROM golang:1.21

WORKDIR /project
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /updater ./cmd/updater

CMD /updater