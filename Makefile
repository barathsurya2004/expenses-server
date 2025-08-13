idlCompile:
	@echo "Compiling IDL files..."
	@protoc --proto_path= ./proto/*.proto \
		   --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative

serve:
	@echo "Starting gRPC server..."
	@go build -o ./services/bin/service ./services/. && ./services/bin/service

run:
	@echo "Running the service..."
	@go build -o ./client/bin/client ./client/. && ./client/bin/client 


file =""
postgresUrl = "postgres://postgres:12345678@localhost:5432/postgres"

migrateCreate:
	@echo "Running database migrations..."
	@migrate create -ext sql -dir db/migrations -seq $(file)

migrateUp:
	@echo "Applying database migrations..."
	@migrate -path db/migrations -database $(postgresUrl) up

migrateDown:
	@echo "Rolling back database migrations..."
	@migrate -path db/migrations -database $(postgresUrl) down