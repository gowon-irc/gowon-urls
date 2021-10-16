FROM golang:alpine as build-env
COPY . /src
WORKDIR /src
RUN go build -o gowon-urls

FROM alpine:3.14.2
WORKDIR /app
COPY --from=build-env /src/gowon-urls /app/
ENTRYPOINT ["./gowon-urls"]
