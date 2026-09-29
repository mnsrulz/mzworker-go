FROM gcr.io/distroless/static-debian12

COPY mzworker-go-linux-amd64 /usr/local/bin/mzworker-go
COPY mzworker-go-linux-amd64 /usr/local/bin/mzworker

ARG BUILD_TIME
ARG GIT_SHA

ENV BUILD_TIME=$BUILD_TIME
ENV GIT_SHA=$GIT_SHA

ENTRYPOINT ["mzworker-go"]
