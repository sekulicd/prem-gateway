#!/bin/bash

set -e

# Get all container IDs connected to the 'prem-gateway' network
CONTAINERS=$(docker ps -aq --filter network=prem-gateway)

# Check if CONTAINERS is empty (no containers on the network)
if [ -z "$CONTAINERS" ]; then
    echo "No containers found on the 'prem-gateway' network."
else
    # Stop all containers
    echo "Stopping containers on 'prem-gateway' network..."
    docker stop "$CONTAINERS"

    # Remove all containers and their anonymous volumes
    echo "Removing containers and cleaning up volumes..."
    docker rm -v "$CONTAINERS"

    echo "Containers stopped and removed. Volumes cleaned."
fi