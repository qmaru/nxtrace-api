FROM golang:alpine AS go-core-builder

RUN apk add upx

RUN CGO_ENABLED=0 go install -ldflags="-w -s -extldflags='static'" -trimpath github.com/nxtrace/NTrace-core@latest \
    && mv $GOPATH/bin/NTrace-core $GOPATH/bin/nexttrace_$(uname -m) \
    && upx --best --lzma $GOPATH/bin/nexttrace_$(uname -m)

FROM golang:alpine AS go-api-builder

WORKDIR /usr/src

COPY . /usr/src/

RUN apk add upx ca-certificates tzdata

RUN gover=`go version | awk '{print $3,$4}'` \
    && sed -i "s#COMMIT_GOVER#$gover#g" utils/version.go \
    && CGO_ENABLED=0 go build -ldflags="-w -s -extldflags='static'" -trimpath -o nxtapi_$(uname -m) \
    && upx --best --lzma nxtapi_$(uname -m)

FROM scratch AS build-api

COPY --from=go-api-builder /usr/src/nxtapi_* /

FROM scratch AS build-core

COPY --from=go-core-builder /go/bin/nexttrace_* /

FROM busybox:uclibc AS tinybox

COPY --from=go-core-builder /go/bin/nexttrace_* /nexttrace
COPY --from=go-api-builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=go-api-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go-api-builder /usr/src/nxtapi_* /nxtapi

RUN rm -fr /bin/*

FROM scratch AS prod

COPY --from=tinybox / /

ENV TRACE_CORE="/nexttrace"

EXPOSE 8080

ENTRYPOINT ["/nxtapi"]
