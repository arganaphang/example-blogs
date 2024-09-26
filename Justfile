set dotenv-load

# Format Golang
format:
	gofumpt -l -w .
	goimports-reviser -rm-unused -set-alias ./...
	golines -w -m 120 *.go

# build -> build application
build:
	go build -o main main.go

# run -> application
run:
	./main

# dev -> run build then run it
dev: 
	watchexec -r -c -e go -- just build run

# health -> Hit Health Check Endpoint
health:
	curl -s http://localhost:8000/healthz | jq

# pgroll-init -> init migration
pgroll-init:
	pgroll init --postgres-url $POSTGRES_URL

# pgroll-start -> start migration
pgroll-start FILENAME:
	pgroll start --postgres-url $POSTGRES_URL {{FILENAME}}

# pgroll-complete -> complete migration
pgroll-complete:
	pgroll complete --postgres-url $POSTGRES_URL

# pgroll-status -> complete migration
pgroll-status:
	pgroll status --postgres-url $POSTGRES_URL

# seed-issue -> seeding issue table
seed-issue:
	go run ./cmd/seeder