#!/bin/bash

# Run the 'make up' command with environment variables
export LETSENCRYPT_PROD=true
export SERVICES=premd,premapp
make up

# Loop to check for 'OK' from curl command
while true; do
  response=$(curl -s http://localhost:8080/ping)
  if [ "$response" == "OK" ]; then
    echo "Received OK. Proceeding to next step."
    break
  else
    echo "Waiting for OK response..."
    sleep 2
  fi
done

# Navigate back to the ./script directory to run 'docker-compose'
cd ./script || { echo "Directory ./script does not exist. Exiting."; exit 1; }

if ! command -v openssl &> /dev/null
then
    sudo apt-get update -qq
    sudo apt-get install -y openssl
fi

# Run the 'docker-compose' command with environment variables
export PREMD_IMAGE
export PREMAPP_IMAGE
docker-compose -f docker-compose-box.yml up -d --build