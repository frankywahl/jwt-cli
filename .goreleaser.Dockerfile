
FROM alpine

ARG BUILD_VERSION
ARG BUILD_DATE
ARG REVISION
ARG SOURCE

LABEL org.opencontainers.image.revision=$REVISION \
  org.opencontainers.image.source=$SOURCE \
  org.opencontainers.image.version=$BUILD_VERSION \
  org.opencontainers.image.created=$BUILD_DATE

COPY jwt /usr/local/bin/.

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser

ENTRYPOINT ["jwt"]
