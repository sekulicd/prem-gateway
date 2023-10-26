PONY: up down runall stopall

## up: run prem-gateway
up:
	export POSTGRES_USER=root; \
	export POSTGRES_PASSWORD=secret; \
	export DNSD_POSTGRES_DB=dnsd-db; \
	export AUTHD_POSTGRES_DB=authd-db; \
	DOCKER_BUILDKIT=0 docker-compose up -d --build

## down: stop prem-gateway
down:
	docker-compose down -v

## runall: run prem-gateway and prem-box
runall:
	chmod +x ./script/run_all.sh
	export PREMD_IMAGE=$(PREMD_IMAGE); \
	export PREMAPP_IMAGE=$(PREMAPP_IMAGE); \
	./script/run_all.sh

## stopall: stop prem-gateway and prem-box
stopall:
	chmod +x ./script/stop_all.sh
	./script/stop_all.sh