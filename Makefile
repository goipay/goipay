dbConnStr = "postgresql://${DATABASE_USER}:${DATABASE_PASS}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}"
migrationsDir = ./sql/migrations
sqlcFile = ./sql/sqlc.yaml

protoPathDir = ./proto
protoGoOutDir = ./internal/pb
protoGoGrpcOutDir = $(protoGoOutDir)

cmdServerDir = ./cmd/server

clean-pb:
	rm -rf $(protoGoOutDir)/*

gen-pb:
	./script/generate_pb.sh -i $(protoPathDir) -o $(protoGoOutDir) v1

run-migrations:
	goose -dir $(migrationsDir) postgres $(dbConnStr) up

reset-migrations:
	goose -dir $(migrationsDir) postgres $(dbConnStr) reset

run-sqlc-gen:
	sqlc -f $(sqlcFile) generate

build:
	go build -o ./bin/server $(cmdServerDir)/main.go

build-debug:
	go build -gcflags=all="-N -l" -o ./bin/server $(cmdServerDir)/main.go

coverage:
	go test -coverprofile=.coverage.out -coverpkg=./internal/handler,./internal/listener,./internal/processor,./internal/util,./internal/db ./... && go tool cover -html=.coverage.out -o .coverage.html

gen-certs:
	rm -rf ./certs/* && bash -x ./script/generate_certs.sh
