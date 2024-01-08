FROM busybox:1.34

COPY shared-init shared-init

ENTRYPOINT [ "./shared-init" ]