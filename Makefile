PONY: build-dns run-dns pg droppg createdb dropdb createtestdb droptestdb recreatedb recreatetestdb pgcreatetestdb psql mig_file mig_up_test mig_up mig_down_test mig_down mig_down_yes vet_db sqlc doc dev up down

##### DNS Daemon #####

## build-dns prem-gateway dns service
build-dns:
	@echo "Building prem-gateway dns service..."
	@export GO111MODULE=on; \
	env go build -tags netgo -ldflags="-s -w" -o bin/dnsd ./dns/cmd/dnsd/main.go

## run-dns runs prem-gateway dns service
run-dns:
	@echo "Running prem-gateway dns service..."
	./bin/dnsd

##### DNS Daemon #####


#### Postgres database ####

## pg: starts postgres db inside docker container
pg:
	docker run --name dnsd-db-pg -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres

## droppg: stop and remove postgres container
droppg:
	docker stop dnsd-db-pg
	docker rm dnsd-db-pg

## createdb: create db inside docker container
createdb:
	docker exec dnsd-db-pg createdb --username=root --owner=root dnsd-db

## dropdb: drops db inside docker container
dropdb:
	docker exec dnsd-db-pg dropdb dnsd-db

## createtestdb: create test db inside docker container
createtestdb:
	docker exec dnsd-db-pg createdb --username=root --owner=root dnsd-db-test

## droptestdb: drops test db inside docker container
droptestdb:
	docker exec dnsd-db-pg dropdb dnsd-db-test

## recreatedb: drop and create main and test db
recreatedb: dropdb createdb droptestdb createtestdb

## recreatetestdb: drop and create test db
recreatetestdb: droptestdb createtestdb

## pgcreatetestdb: starts docker container and creates test db, used in CI
pgcreatetestdb:
	chmod u+x ./script/create_testdb
	./script/create_testdb

## psql: connects to postgres terminal running inside docker container
psql:
	docker exec -it dnsd-db-pg psql -U root -d dnsd-db


## mig_file: creates pg migration file(eg. make FILE=init mig_file)
mig_file:
	@migrate create -ext sql -dir ./dns/internal/infrastructure/storage/pg/migration/ $(FILE)

## mig_up_test: creates test db schema
mig_up_test:
	@echo "creating db schema..."
	@migrate -database "postgres://root:secret@localhost:5432/dnsd-db-test?sslmode=disable" -path ./dns/internal/infrastructure/storage/pg/migration/ up

## mig_up: creates db schema
mig_up:
	@echo "creating db schema..."
	@migrate -database "postgres://root:secret@localhost:5432/dnsd-db?sslmode=disable" -path ./dns/internal/infrastructure/storage/pg/migration/ up

## mig_down_test: apply down migration on test db
mig_down_test:
	@echo "migration down on test db..."
	@migrate -database "postgres://root:secret@localhost:5432/dnsd-db-test?sslmode=disable" -path ./dns/internal/infrastructure/storage/pg/migration/ down

## mig_down: apply down migration
mig_down:
	@echo "migration down..."
	@migrate -database "postgres://root:secret@localhost:5432/dnsd-db?sslmode=disable" -path ./dns/internal/infrastructure/storage/pg/migration/ down

## mig_down_yes: apply down migration without prompt
mig_down_yes:
	@echo "migration down..."
	@"yes" | migrate -database "postgres://root:secret@localhost:5432/dnsd-db?sslmode=disable" -path ./dns/internal/infrastructure/storage/pg/migration/ down

## vet_db: check if mig_up and mig_down are ok
vet_db: recreatedb mig_up mig_down_yes
	@echo "vet db migration scripts..."

## sqlc: gen sql
sqlc:
	@echo "gen sql..."
	cd ./dns/internal/infrastructure/storage/pg; sqlc generate

#### Postgres database ####


#### Swagger doc ####

## doc: generate swagger doc
doc:
	@echo "generating swagger doc..."
	swag init -g ./dns/cmd/dnsd/main.go -o ./dns/docs

#### Swagger doc ####

## dev-dns: run dnsd and postgres
dev-dns:
	export POSTGRES_USER=root; \
	export POSTGRES_PASSWORD=secret; \
	export POSTGRES_DB=dnsd-db; \
	cd ./dns; \
	DOCKER_BUILDKIT=0 docker-compose up -d --build

## up: run prem-gateway
up:
	export POSTGRES_USER=root; \
	export POSTGRES_PASSWORD=secret; \
	export POSTGRES_DB=dnsd-db; \
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

#### Go lint ####

## vetdnsd: run go vet on dnsd
vetdnsd:
	@echo "go vet dnsd..."
	@cd dns && go vet ./...

#### Go lint ####

#### Go mock ####

## mockdnsd: generater mocks
mockdnsd:
	cd ./dns/internal/core/port/; \
    mockery --name=ControllerdWrapper --structname=MockControllerdWrapper \
	--output=./ --outpkg=port --filename=controllerd_wrapper_mock.go --inpackage; \
	mockery --name=IpService --structname=MockIpService \
	--output=./ --outpkg=port --filename=ip_service_mock.go --inpackage;

#### Go mock ####