FROM ghcr.io/osgeo/gdal:alpine-small-latest AS builder
RUN apk add --no-cache go gcc musl-dev pkgconfig coreutils libseccomp-dev
WORKDIR /app

COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download

COPY min min
COPY scripts scripts
COPY internal internal
COPY main.go .

RUN ./scripts/build.sh

FROM ghcr.io/osgeo/gdal:alpine-small-latest
RUN apk add --no-cache libseccomp && \
    rm -rf \
        /var/cache/apk/*  \
        /lib/apk \
        /usr/bin \
        /bin \
        /var \
        /media \
        /sbin \
        /usr/sbin \
        /usr/share \
        /usr/include

COPY --from=builder --chown=1000:1000 /app/bin/topography /usr/local/bin/topography

USER 1000:1000
ENTRYPOINT ["topography", "--server"]
