FROM ghcr.io/osgeo/gdal:alpine-small-3.12.2 AS builder

RUN apk add --no-cache go gcc musl-dev pkgconfig coreutils
WORKDIR /app

COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download

COPY min min
COPY scripts scripts
COPY internal internal
COPY main.go .

RUN ./scripts/build.sh

FROM ghcr.io/osgeo/gdal:alpine-small-3.12.2
WORKDIR /app

COPY --from=builder /app/bin/topography /usr/local/bin/topography
ENTRYPOINT ["topography", "--server"]