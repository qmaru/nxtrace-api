# nxtrace-api-docker

nxtrace tiny api (web/mqtt)

## How to use

### env

|NAME|DEFAULT|
|:-:|:-:|
|TRACE_CORE|/usr/bin/nexttrace|
|TRACE_WEB_HOST|127.0.0.1|
|TRACE_WEB_PORT|8080|
|TRACE_MQTT_HOST|127.0.0.1|
|TRACE_MQTT_PORT|1883|
|TRACE_MQTT_USERNAME|qmaru|
|TRACE_MQTT_PASSWORD|123456|
|TRACE_MQTT_TOPIC|trace/data|
|TRACE_MQTT_CLIENT|qmeta-pub-xxx|
|TRACE_MQTT_WITHTLS|false|

### server

```shell
podman run -d \
    --name nxtapi \
    --restart=always \
    --privileged=true \
    -e TRACE_MQTT_HOST="mqtt.broker.com" \
    -e TRACE_MQTT_PORT="443" \
    -e TRACE_MQTT_USERNAME="username" \
    -e TRACE_MQTT_PASSWORD="password" \
    -e TRACE_MQTT_WITHTLS="true" \
    ghcr.io/qmaru/nxtrace:go
```

## Credits

[NTrace-core](https://github.com/nxtrace/NTrace-core)
