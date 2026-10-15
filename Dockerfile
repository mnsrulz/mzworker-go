FROM gcr.io/distroless/cc-debian13:nonroot

ARG TARGETARCH

COPY --chmod=755 mzworker-go-linux-${TARGETARCH} /usr/local/bin/mzworker-go

ENTRYPOINT ["/usr/local/bin/mzworker-go", "serve"]
