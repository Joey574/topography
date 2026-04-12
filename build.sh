#!/bin/sh
start_time=$(date +%s.%N)

LDFLAGS="-s -w -linkmode=external -extldflags '-Wl,-z,relro,-z,now,-z,noexecstack'"
GCFLAGS="-c=8 -dwarf=false -trimpath"
BUILDFLAGS="-tags=netgo,osusergo -buildmode=pie -buildvcs=false -trimpath"

OUTPUT="./bin/topography"

go build $BUILDFLAGS -ldflags="$LDFLAGS" -gcflags="$GCFLAGS" -o $OUTPUT
strip $OUTPUT

file_size=$(stat -c %s $OUTPUT)
size_human=$(numfmt --to=iec --suffix=B "$file_size")
end_time=$(date +%s.%N)
elapsed=$(echo "$end_time - $start_time" | bc)

printf "Build completed in %.2f seconds (%s)\n" "$elapsed" "$size_human"
