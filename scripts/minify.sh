#!/bin/sh

OLD_HASH=$(find ./min/* -type f -exec sha256sum {} + | LC_ALL=C sort | sha256sum)

rm -rf ./min
mkdir -p ./min
cp -r ./src/* ./min
if command -v minify >/dev/null 2>&1; then
    minify -q -r -i ./min
else
    echo "[-] WARNING: 'minify' command not found, this is NOT a fatal error"
fi

NEW_HASH=$(find ./min/* -type f -exec sha256sum {} + | LC_ALL=C sort | sha256sum)

printf "[OLD] [%s]\n[NEW] [%s]\n" "${OLD_HASH}" "${NEW_HASH}"
