#!/bin/bash
BASIC_AUTH_USER="admin"
BASIC_AUTH_PASS=$(openssl rand -base64 4)
ROOT_KEY=$(openssl rand -base64 8)
# Run the 'make up' command with environment variables

PREM_GATEWAY_AUTH_ROOT_API_KEY=$ROOT_KEY
PREM_GATEWAY_AUTH_ADMIN_USER=$BASIC_AUTH_USER
PREM_GATEWAY_AUTH_ADMIN_PASS=$BASIC_AUTH_PASS
export PREM_GATEWAY_AUTH_ROOT_API_KEY
export PREM_GATEWAY_AUTH_ADMIN_USER
export PREM_GATEWAY_AUTH_ADMIN_PASS
export LETSENCRYPT_PROD=true
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

echo "Basic auth user: $BASIC_AUTH_USER"
echo "Basic auth pass: $BASIC_AUTH_PASS"