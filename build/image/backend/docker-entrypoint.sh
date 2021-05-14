#!/bin/bash
set -e

serverMode="common-backend"
if [ "$FABRIC_BAAS_SERVER_MODE" ]; then
    serverMode=$FABRIC_BAAS_SERVER_MODE
fi

if [ "$1" == "commonRun" ]; then
    # Migrate database and start the BaaS Backend server
    common-backend server migrate -m $serverMode && common-backend server start -m $serverMode
    exit
fi

exec "$@"