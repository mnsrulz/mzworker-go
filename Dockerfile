FROM gcr.io/distroless/static-debian12

COPY mzworker-go-linux-amd64 /usr/local/bin/mzworker-go
COPY mzworker-go-linux-amd64 /usr/local/bin/mzworker

ENTRYPOINT ["mzworker-go"]
