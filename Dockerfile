FROM golang:alpine as go-core-builder

RUN apk add upx

RUN CGO_ENABLED=0 go install -ldflags="-w -s -extldflags='static'" -trimpath github.com/nxtrace/NTrace-core@latest \
    && mv $GOPATH/bin/NTrace-core $GOPATH/bin/nexttrace \
    && upx --best --lzma $GOPATH/bin/nexttrace

FROM golang:alpine as go-api-builder

WORKDIR /usr/src

COPY . /usr/src/

RUN apk add upx ca-certificates tzdata

RUN gover=`go version | awk '{print $3,$4}'` \
    && sed -i "s#COMMIT_GOVER#$gover#g" utils/version.go \
    && CGO_ENABLED=0 go build -ldflags="-w -s -extldflags='static'" -trimpath -o app \
    && upx --best --lzma app

FROM scratch as build-api

COPY --from=go-api-builder /usr/src/app /nxtapi

FROM scratch as build-core

COPY --from=go-core-builder /go/bin/nexttrace /nexttrace

FROM busybox:uclibc as tinybox

RUN rm -fr /bin/*

COPY --from=go-core-builder /go/bin/nexttrace /nexttrace
COPY --from=go-api-builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=go-api-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go-api-builder /usr/src/app /nxtapi

FROM scratch as prod

COPY --from=tinybox / /

ENV TRACE_CORE="/nexttrace"

EXPOSE 8080

ENTRYPOINT ["/nxtapi"]
