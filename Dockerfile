FROM alpine

#assumes distest was pre-built for the
COPY disktest /var/opt/disktest/disktest

VOLUME [ "/data" ]

ENV GENERATE=y
ENV VERIFY=mem
ENV SIZE=1GB
ENV EXTRA_FLAGS=

ENTRYPOINT [ "sh", "-c", "/var/opt/disktest/disktest \
    --generate $GENERATE \
    --verify $VERIFY \
    --size $SIZE \
    $EXTRA_FLAGS \
    /data" ]