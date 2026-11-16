# Minimal glibc runtime with a current glibc (2.44 at the time of writing).
# distroless cc-debian13 was replaced because it ships glibc 2.41, while the
# ubuntu-26.04 CI runners link against glibc 2.43. Pinned by digest because
# the glibc-dynamic tag rolls forward; bump the digest deliberately.
FROM cgr.dev/chainguard/glibc-dynamic:latest@sha256:a5e9991904f2034c4b39d27e3121ce7eba5a667ead9b5f3a96bfbda259be4f29

ARG TARGETARCH

COPY --chmod=755 mzworker-go-linux-${TARGETARCH} /usr/local/bin/mzworker-go

ENTRYPOINT ["/usr/local/bin/mzworker-go", "serve"]
