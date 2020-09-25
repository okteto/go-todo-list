FROM golang:buster as builder

WORKDIR /app
ADD . .
RUN go build -o app

##########################

FROM debian:buster as prod

WORKDIR /app
COPY --from=builder /app/app /app/app
COPY --from=builder /app/static /app/static
EXPOSE 8080
CMD ["./app"]