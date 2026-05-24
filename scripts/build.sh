#!/bin/sh

tags=""
while getopts "t:" opt; do
  case $opt in
    t) tags="${tags},${OPTARG}" ;;
    \?) echo "Invalid option -$OPTARG" >&2; exit 1 ;;
  esac
done
tags=${tags#,}

C_FLAGS="-O3 -march=x86-64-v3 -flto -U_FORTIFY_SOURCE -D_FORTIFY_SOURCE=3 -fstack-protector-strong"
CPP_FLAGS="-O3 -march=x86-64-v3 -flto -U_FORTIFY_SOURCE -D_FORTIFY_SOURCE=3"

CGO_LDFLAGS="-O3 -flto -Wl,-z,relro,-z,now,-z,noexecstack"
LDFLAGS="-s -w -linkmode=external -extldflags '$CGO_LDFLAGS'"

BUILDFLAGS="-tags=netgo,osusergo,${tags} -buildmode=pie -buildvcs=false -trimpath"

OUTPUT="./bin/topography"

CGO_ENABLED=1 CGO_CFLAGS="$C_FLAGS" CGO_CPPFLAGS="$CPP_FLAGS" GOAMD64=v3 \
    go build \
    $BUILDFLAGS \
    -ldflags="$LDFLAGS" \
    -o $OUTPUT
