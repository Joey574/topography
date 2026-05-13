#!/bin/sh
C_FLAGS="-O3 -march=native -mtune=native -flto -U_FORTIFY_SOURCE -D_FORTIFY_SOURCE=3 -fstack-protector-strong"
CPP_FLAGS="-O3 -march=native -mtune=native -flto -U_FORTIFY_SOURCE -D_FORTIFY_SOURCE=3"

CGO_LDFLAGS="-O3 -flto -Wl,-z,relro,-z,now,-z,noexecstack"
LDFLAGS="-s -w -linkmode=external -extldflags '$CGO_LDFLAGS'"

BUILDFLAGS="-tags=netgo,osusergo -buildmode=pie -buildvcs=false -trimpath"

OUTPUT="./bin/topography"

CGO_ENABLED=1 CGO_CFLAGS="$C_FLAGS" CGO_CPPFLAGS="$CPP_FLAGS" \
    go build \
    $BUILDFLAGS \
    -ldflags="$LDFLAGS" \
    -o $OUTPUT
