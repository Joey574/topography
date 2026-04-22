FROM golang:1.26-alpine AS builder

# pull in gdal
RUN sed -i -e 's/v[0-9]\.[0-9]\+/edge/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache gcc musl-dev pkgconfig gdal-dev


WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN sh ./scripts/build.sh

FROM alpine:edge
RUN apk update && \
    apk add --no-cache gdal && \
    rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=builder /app/bin/topography /usr/local/bin/topography

# Execute the binary
ENTRYPOINT ["topography", "--server"]