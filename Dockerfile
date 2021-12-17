FROM golang:1.16 as build
RUN apk add make git
ADD . /go/src/github.com/kube-queue/et-operator-extension

WORKDIR /go/src/github.com/kube-queue/et-operator-extension
RUN make

FROM alpine:3.12
COPY --from=build /go/src/github.com/kube-queue/et-operator-extension/bin/et-operator-extension /usr/bin/et-operator-extension
RUN chmod +x /usr/bin/et-operator-extension
ENTRYPOINT ["/usr/bin/et-operator-extension"]
