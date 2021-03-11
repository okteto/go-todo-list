FROM okteto/golang:1 as dev

RUN apt-get update && apt-get install -y postgresql-client

RUN curl -L curl -L https://github.com/rberrelleza/proxy/releases/download/0.1.1/proxy-0.1.1-linux-amd64.tar.gz | tar -zx -C /usr/local/bin

WORKDIR /usr/src/app
ADD . .
RUN go build -o app

##########################

FROM debian:buster as prod

WORKDIR /app
COPY --from=builder /usr/src/app/app /app/app
COPY --from=builder /usr/src/app/static /app/static

EXPOSE 8080
CMD ["./app"]