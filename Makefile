.PHONY: help up start-temporal start-workflow-server

help:
	@echo "Usage: make <target>"
	@echo "targets:"
	@echo "  up                      - Starts Temporal and the workflow server in the background"
	@echo "  start-temporal          - Starts the Temporal development server"
	@echo "  start-workflow-server   - Starts the workflow server"

up:
	@echo "Starting Temporal server in background..."
	@temporal server start-dev &
	@echo "Starting workflow server in background..."
	@go run cmd/server/main.go &

start-temporal:
	@echo "Starting Temporal development server..."
	@temporal server start-dev

start-workflow-server:
	@echo "Starting workflow server..."
	@go run cmd/server/main.go
