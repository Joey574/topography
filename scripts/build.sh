#!/bin/sh
set -e
LDFLAGS="-s -w -linkmode=external -extldflags '-Wl,-z,relro,-z,now,-z,noexecstack'"
GCFLAGS="-c=8 -dwarf=false -trimpath"
BUILDFLAGS="-tags=netgo,osusergo -buildmode=pie -buildvcs=false -trimpath"
C_FLAGS="-O3 -march=native -mtune=native -fstack-protector-strong"
CPP_FLAGS="-O3 -march=native -mtune=native -D_FORTIFY_SOURCE=2"

OUTPUT="./bin/topography"

CGO_ENABLED=1 CGO_CFLAGS="$C_FLAGS" CGO_CPPFLAGS="$CPP_FLAGS" go build $BUILDFLAGS -ldflags="$LDFLAGS" -gcflags="$GCFLAGS" -o $OUTPUT
strip $OUTPUT
