#!/bin/sh

rm -rf ./min
mkdir -p ./min
cp -r ./src/* ./min
if command -v minify >/dev/null 2>&1; then
    minify -q -r -i ./min
else
    echo "[-] WARNING: 'minify' command not found, this is NOT a fatal error"
fi
