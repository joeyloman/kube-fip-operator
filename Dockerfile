FROM docker.io/golang:1.20-alpine3.17 as builder
RUN mkdir /src /deps
RUN apk update && apk add git build-base binutils-gold
WORKDIR /deps
ADD go.mod /deps
RUN go mod download
ADD / /src
WORKDIR /src/pkg/cmd/kube-fip-operator
RUN go build -o kube-fip-operator .
FROM docker.io/alpine:3.17
RUN adduser -S -D -H -h /app kubefip
USER kubefip
COPY --from=builder /src/pkg/cmd/kube-fip-operator/kube-fip-operator /app/
WORKDIR /app
ENTRYPOINT ["./kube-fip-operator"]
