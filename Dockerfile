FROM golang:alpine as go-builder

WORKDIR /usr/src

COPY . /usr/src/

RUN apk add upx ca-certificates tzdata

RUN CGO_ENABLED=0 go install -ldflags="-w -s -extldflags='static'" -trimpath github.com/nxtrace/NTrace-core@latest \
    && mv $GOPATH/bin/NTrace-core $GOPATH/bin/nexttrace \
    && upx --best --lzma $GOPATH/bin/nexttrace

RUN CGO_ENABLED=0 go build -ldflags="-w -s -extldflags='static'" -trimpath -o app \
    && upx --best --lzma app

FROM scratch as build-api

COPY --from=go-builder /usr/src/app /nxtapi

FROM scratch as build-core

COPY --from=go-builder /go/bin/nexttrace /nexttrace

FROM busybox:uclibc as tinybox

RUN rm -fr /bin/*

COPY --from=go-builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go-builder /go/bin/nexttrace /nexttrace
COPY --from=go-builder /usr/src/app /nxtapi

FROM scratch as prod

COPY --from=tinybox / /

ENV TRACE_CORE="/nexttrace"

EXPOSE 8080

ENTRYPOINT ["/nxtapi"]
