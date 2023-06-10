FROM busybox:1.34

COPY ambient-init ambient-init

ENTRYPOINT [ "./ambient-init" ]