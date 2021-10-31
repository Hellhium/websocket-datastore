FROM golang:latest as builder
LABEL maintainer="Jemy SCHNEPP <dev@leethium.fr>"

WORKDIR /app

COPY . .
RUN go build -o wsapi


FROM alpine:latest

RUN apk add --no-cache ca-certificates musl
WORKDIR /root
RUN mkdir /root/data
VOLUME /root/data

COPY --from=builder /app/wsapi .

EXPOSE 8080

# Command to run the executable
ENTRYPOINT []
CMD ["./wsapi"]
