PONY: build run pg droppg createdb dropdb createtestdb droptestdb recreatedb recreatetestdb pgcreatetestdb psql mig_file mig_up_test mig_up mig_down_test mig_down mig_down_yes vet_db sqlc doc up vet

## build authd
build:
	@echo "Building authd..."
	@export GO111MODULE=on; \
	env go build -tags netgo -ldflags="-s -w" -o bin/authd ./cmd/authd/main.go

## run authd
run:
	@echo "Running dns service..."
	./bin/authd

#### Postgres database ####

## pg: starts postgres db inside docker container
pg:
	docker run --name authd-db-pg -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres

## droppg: stop and remove postgres container
droppg:
	docker stop authd-db-pg
	docker rm authd-db-pg

## createdb: create db inside docker container
createdb:
	docker exec authd-db-pg createdb --username=root --owner=root authd-db

## dropdb: drops db inside docker container
dropdb:
	docker exec authd-db-pg dropdb authd-db

## createtestdb: create test db inside docker container
createtestdb:
	docker exec authd-db-pg createdb --username=root --owner=root authd-db-test

## droptestdb: drops test db inside docker container
droptestdb:
	docker exec authd-db-pg dropdb authd-db-test

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
	docker exec -it authd-db-pg psql -U root -d authd-db


## mig_file: creates pg migration file(eg. make FILE=init mig_file)
mig_file:
	@migrate create -ext sql -dir ./internal/infrastructure/storage/pg/migration/ $(FILE)

## mig_up_test: creates test db schema
mig_up_test:
	@echo "creating db schema..."
	@migrate -database "postgres://root:secret@localhost:5432/authd-db-test?sslmode=disable" -path ./internal/infrastructure/storage/pg/migration/ up

## mig_up: creates db schema
mig_up:
	@echo "creating db schema..."
	@migrate -database "postgres://root:secret@localhost:5432/authd-db?sslmode=disable" -path ./internal/infrastructure/storage/pg/migration/ up

## mig_down_test: apply down migration on test db
mig_down_test:
	@echo "migration down on test db..."
	@migrate -database "postgres://root:secret@localhost:5432/authd-db-test?sslmode=disable" -path ./internal/infrastructure/storage/pg/migration/ down

## mig_down: apply down migration
mig_down:
	@echo "migration down..."
	@migrate -database "postgres://root:secret@localhost:5432/authd-db?sslmode=disable" -path ./internal/infrastructure/storage/pg/migration/ down

## mig_down_yes: apply down migration without prompt
mig_down_yes:
	@echo "migration down..."
	@"yes" | migrate -database "postgres://root:secret@localhost:5432/authd-db?sslmode=disable" -path ./internal/infrastructure/storage/pg/migration/ down

## vet_db: check if mig_up and mig_down are ok
vet_db: recreatedb mig_up mig_down_yes
	@echo "vet db migration scripts..."

## sqlc: gen sql
sqlc:
	@echo "gen sql..."
	cd ./internal/infrastructure/storage/pg; sqlc generate

#### Postgres database ####


#### Swagger doc ####

## doc: generate swagger doc
doc:
	@echo "generating swagger doc..."
	swag init -g ./cmd/authd/main.go -o ./docs

#### Swagger doc ####

## up: run authd and postgres
up:
	export POSTGRES_USER=root; \
	export POSTGRES_PASSWORD=secret; \
	export POSTGRES_DB=authd-db; \
	DOCKER_BUILDKIT=0 docker-compose up -d --build

#### Go lint ####

## vet: run go vet on authd
vet:
	@echo "go vet authd..."
	go vet ./...

#### Go lint ####

#### Go mock ####

## mock: generater mocks
mock:
	cd ./internal/core/domain/; \
	mockery --name=ApiKeyRepository --structname=MockApiKeyRepository \
	--output=./ --outpkg=port --filename=api_key_repository_mock.go --inpackage; \
	mockery --name=RepositoryService --structname=MockRepositoryService \
    	--output=./ --outpkg=port --filename=repository_service_mock.go --inpackage;