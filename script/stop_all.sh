#!/bin/bash

# Ensure the script stops on first error
set -e

# Get all container IDs connected to the 'prem-gateway' network
CONTAINERS=($(docker ps -aq --filter network=prem-gateway))

# Check if CONTAINERS is empty (no containers on the network)
if [ ${#CONTAINERS[@]} -eq 0 ]; then
    echo "No containers found on the 'prem-gateway' network."
else
    PGVOLUME=($(docker inspect dnsd-db-pg | jq -r '.[0].HostConfig.Mounts[0].Source'))

    # Stop all containers
    echo "Stopping containers on 'prem-gateway' network..."
    for CONTAINER in "${CONTAINERS[@]}"; do
        docker stop "$CONTAINER"
    done

    # Remove all containers and their anonymous volumes
    echo "Removing containers and cleaning up volumes..."
    for CONTAINER in "${CONTAINERS[@]}"; do
        docker rm "$CONTAINER"
    done

    # Remove dnsd-db-pg volume
    echo "Remove dnsd-db-pg volume"
    docker volume rm "$PGVOLUME"

    echo "Containers stopped and removed. Volumes cleaned."
fi