FROM gcr.io/distroless/base-debian13:nonroot

ARG TARGETARCH

COPY mzworker-go-linux-${TARGETARCH} /usr/local/bin/mzworker-go

ENTRYPOINT ["/usr/local/bin/mzworker-go"]
