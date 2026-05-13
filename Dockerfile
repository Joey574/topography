FROM ghcr.io/osgeo/gdal:alpine-small-latest AS builder
RUN apk add --no-cache go gcc musl-dev pkgconfig coreutils libseccomp-dev
WORKDIR /app

COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY min min
COPY scripts scripts
COPY internal internal
COPY main.go .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    ./scripts/build.sh

FROM ghcr.io/osgeo/gdal:alpine-small-latest
RUN apk add --no-cache libseccomp && \
   addgroup -S server && \
   adduser -S server -G server
USER server

COPY --from=builder --chown=server /app/bin/topography /usr/local/bin/topography
ENTRYPOINT ["topography", "--server"]
