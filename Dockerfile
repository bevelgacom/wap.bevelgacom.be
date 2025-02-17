FROM golang:1.23-alpine as build

RUN apk add --no-cache imagemagick-dev gcc musl-dev pkgconfig

ENV CGO_ENABLED=1

COPY ./ /go/src/github.com/bevelgacom/wap.bevelgacom.be

WORKDIR /go/src/github.com/bevelgacom/wap.bevelgacom.be

RUN go build -o server ./

FROM alpine:edge

RUN apk add --no-cache ca-certificates tzdata

RUN mkdir /opt/wap.bevelgacom.be
WORKDIR /opt/wap.bevelgacom.be

COPY --from=build /go/src/github.com/bevelgacom/wap.bevelgacom.be/hafas-stations.ndjson /opt/wap.bevelgacom.be/
COPY --from=build /go/src/github.com/bevelgacom/wap.bevelgacom.be/server /usr/local/bin
COPY --from=build /go/src/github.com/bevelgacom/wap.bevelgacom.be/static /opt/wap.bevelgacom.be/static

ENTRYPOINT [ "server" ]
