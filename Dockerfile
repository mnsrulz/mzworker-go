FROM gcr.io/distroless/static-debian12

COPY mzworker-linux-amd64 /usr/local/bin/mzworker

ENTRYPOINT ["mzworker"]
