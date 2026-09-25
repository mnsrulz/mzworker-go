FROM gcr.io/distroless/static-debian12

COPY mzworker-go-linux-amd64 /usr/local/bin/mzworker-go
RUN ln -sf /usr/local/bin/mzworker-go /usr/local/bin/mzworker

ENTRYPOINT ["mzworker-go"]
