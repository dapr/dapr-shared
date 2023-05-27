FROM busybox:latest

COPY main main

ENTRYPOINT [ "./main" ]