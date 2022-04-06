FROM docker.io/golang:1.17-alpine3.14 as builder
RUN mkdir /src /deps
RUN apk update && apk add git build-base
WORKDIR /deps
ADD go.mod /deps
RUN go mod download
ADD / /src
WORKDIR /src/pkg/cmd/kube-fip-operator
RUN go build -o kube-fip-operator .
FROM docker.io/alpine:3.14
RUN adduser -S -D -H -h /app kubefip
USER kubefip
COPY --from=builder /src/pkg/cmd/kube-fip-operator/kube-fip-operator /app/
WORKDIR /app
ENTRYPOINT ["./kube-fip-operator"]
