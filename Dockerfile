FROM docker.io/golang:1.23-alpine3.20 as builder
RUN mkdir /src /deps
RUN apk update && apk add git build-base binutils-gold
WORKDIR /deps
ADD go.mod /deps
RUN go mod download
ADD / /src
WORKDIR /src
RUN go build -o kube-fip-operator .
FROM docker.io/alpine:3.20
RUN adduser -S -D -H -h /app kubefip
USER kubefip
COPY --from=builder /src/kube-fip-operator /app/
WORKDIR /app
ENTRYPOINT ["./kube-fip-operator"]