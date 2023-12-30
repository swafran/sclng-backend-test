FROM golang:1.21

WORKDIR /project
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /front ./cmd/front

CMD /front
