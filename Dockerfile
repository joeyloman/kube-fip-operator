FROM docker.io/golang:1.24-alpine3.21 AS builder
RUN mkdir /src /deps
RUN apk update && apk add git build-base binutils-gold
WORKDIR /deps
ADD go.mod /deps
RUN go mod download
ADD / /src
WORKDIR /src
RUN go build -mod vendor -o kube-fip-operator .
FROM docker.io/alpine:3.21
RUN adduser -S -D -H -h /app kubefip
USER kubefip
COPY --from=builder /src/kube-fip-operator /app/
WORKDIR /app
ENTRYPOINT ["./kube-fip-operator"]