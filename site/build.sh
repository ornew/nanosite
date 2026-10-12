#!/bin/sh
set -eu
cd "$(dirname "$0")"
exec go run .. -repo .. -config config.json -out dist "$@"
