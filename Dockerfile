FROM golang:buster as builder
LABEL maintainer="Jemy SCHNEPP <dev@leethium.fr>"

WORKDIR /app

COPY . .
RUN make


FROM alpine:latest

RUN apk add --no-cache ca-certificates musl libc6-compat
WORKDIR /root
RUN mkdir /root/data
VOLUME /root/data

COPY --from=builder /app/build .

EXPOSE 8080

# Command to run the executable
ENTRYPOINT []
CMD ["./bin/wsapi"]
