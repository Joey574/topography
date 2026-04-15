#!/bin/sh
start_time=$(date +%s.%N)

LDFLAGS="-s -w -linkmode=external -extldflags '-Wl,-z,relro,-z,now,-z,noexecstack'"
GCFLAGS="-c=8 -dwarf=false -trimpath"
BUILDFLAGS="-tags=netgo,osusergo -buildmode=pie -buildvcs=false -trimpath"

OUTPUT="./bin/topography"

if command -v minify >/dev/null 2>&1; then
    minify -q ./src/style.css -o ./static/css/style.css
    minify -q ./src/script.js -o ./static/js/script.js
    minify -q ./src/index.html -o ./templates/index.html
else
    echo "[-] WARNING: 'minify' command not found, this is NOT a fatal error"
    cp ./src/style.css ./static/css/style.css
    cp ./src/script.js ./static/js/script.js
    cp ./src/index.html ./templates/index.html
fi

go build $BUILDFLAGS -ldflags="$LDFLAGS" -gcflags="$GCFLAGS" -o $OUTPUT
strip $OUTPUT

file_size=$(stat -c %s $OUTPUT)
size_human=$(numfmt --to=iec --suffix=B "$file_size")
end_time=$(date +%s.%N)
elapsed=$(echo "$end_time - $start_time" | bc)

printf "Build completed in %.2f seconds (%s)\n" "$elapsed" "$size_human"
