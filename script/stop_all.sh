#!/bin/bash

make down

cd ./script

export PREMD_IMAGE
export PREMAPP_IMAGE
docker-compose -f docker-compose-box.yml down -v