#!/bin/sh
set -e

echo "================================================"
echo "TigerBeetle Initialization Script"
echo "================================================"

DATA_FILE="/data/cluster_0_replica_0.tigerbeetle"

echo "Checking for data file: $DATA_FILE"

if [ ! -f "$DATA_FILE" ]; then
    echo "Data file not found. Initializing TigerBeetle data file..."
    echo "Running: /tigerbeetle format --cluster=0 --replica=0 --replica-count=1 $DATA_FILE"

    # Format the data file
    /tigerbeetle format --cluster=0 --replica=0 --replica-count=1 "$DATA_FILE"

    if [ $? -eq 0 ]; then
        echo "Data file initialized successfully!"
    else
        echo "ERROR: Failed to initialize data file"
        exit 1
    fi
else
    echo "Data file already exists. Skipping initialization."
fi

echo "================================================"
echo "Starting TigerBeetle server..."
echo "Cluster ID: 0"
echo "Replica ID: 0"
echo "Listening on: 0.0.0.0:3000"
echo "Cache Grid: 256MiB"
echo "Data file: $DATA_FILE"
echo "================================================"

# Start TigerBeetle server (this replaces the current process)
exec /tigerbeetle start --addresses=0.0.0.0:3000 --cache-grid=256MiB "$DATA_FILE"
