#!/bin/bash
set -euo pipefail

cat >/tmp/request-body.txt
trap "rm -f /tmp/request-body.txt" EXIT
echo "
REMOTE_ADDR=${REMOTE_ADDR}
REQUEST_METHOD=${REQUEST_METHOD}
REQUEST_URI=${REQUEST_URI}
REQUEST_HOST=${REQUEST_HOST}
REQUEST_PROTO=${REQUEST_PROTO}
REQUEST_PATH=${REQUEST_PATH}

REQUEST_HEADERS=[${REQUEST_HEADERS}]

REQUEST_PARAMS=[${REQUEST_PARAMS}]

CONTENT_LENGTH=${CONTENT_LENGTH}
REQUEST_BODY=[$(cat /tmp/request-body.txt)]
"

echo "file list:"
ls -lh
