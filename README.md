# nxtrace-api-docker

nxtrace tiny api (web/mqtt)

## How to use

### env

|NAME|DEFAULT|
|:-:|:-:|
|TRACE_DEBUG|false|
|TRACE_TIMEOUT|120|
|TRACE_CORE|/usr/bin/nexttrace|
|TRACE_WEB_HOST|127.0.0.1|
|TRACE_WEB_PORT|8080|
|TRACE_MQTT_TYPE|ws|
|TRACE_MQTT_HOST|127.0.0.1|
|TRACE_MQTT_PORT|1883|
|TRACE_MQTT_USERNAME|qmaru|
|TRACE_MQTT_PASSWORD|123456|
|TRACE_MQTT_TOPIC|trace/data|
|TRACE_MQTT_QOS|0|
|TRACE_MQTT_RETAIN|false|
|TRACE_MQTT_CLEANSTART|false|
|TRACE_MQTT_CLIENT|qmeta-pub-xxx|
|TRACE_MQTT_WITHTLS|false|

### server

```shell
docker run \
    --name nxtapi \
    --rm \
    --net=host \
    --privileged=true \
    -e TRACE_MQTT_TYPE="tcp" \
    -e TRACE_MQTT_HOST=127.0.0.1 \
    -e TRACE_MQTT_PORT="1883" \
    -e TRACE_MQTT_USERNAME=qmaru \
    -e TRACE_MQTT_PASSWORD=123456 \
    -e TRACE_MQTT_TOPIC="trace/data" \
    -e TRACE_MQTT_CLIENT=qmeta-pub-xxx \
    -e TRACE_MQTT_WITHTLS="false" \
    ghcr.io/qmaru/nxtrace:latest mqtt
```

## Credits

[NTrace-core](https://github.com/nxtrace/NTrace-core)
