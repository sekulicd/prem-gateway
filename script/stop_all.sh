#!/bin/bash

# Get all container IDs connected to the 'prem-gateway' network
CONTAINERS=($(docker ps -aq --filter network=prem-gateway))

# Check if CONTAINERS is empty (no containers on the network)
if [ ${#CONTAINERS[@]} -eq 0 ]; then
    echo "No containers found on the 'prem-gateway' network."
else
    # Stop all containers
    echo "Stopping containers on 'prem-gateway' network..."
    for CONTAINER in "${CONTAINERS[@]}"; do
        docker stop "$CONTAINER"
    done

    # Remove all containers and their anonymous volumes
    echo "Removing containers and cleaning up volumes..."
    for CONTAINER in "${CONTAINERS[@]}"; do
        docker rm -v "$CONTAINER"
    done

    echo "Containers stopped and removed. Volumes cleaned."
fi
