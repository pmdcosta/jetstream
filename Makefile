default: help
.PHONY: help dev start stop nats-create-stream nats-list-streams nats-describe-stream nats-backup-stream nats-edit-stream nats-clear-stream


# LOCAL DEVELOPMENT
dev: ## start application with local configuration.
	go run cmd/producer/main.go

start: ## start local docker containers.
	docker-compose up -d

stop: ## stop local docker containers.
	docker-compose down


# NATS COMMANDS
nats-create-stream: ## creates a file backed stream with unlimited size and 3 replicas.
	@nats str add ORDERS --subjects "ORDERS.*" --ack --max-msgs=-1 --max-bytes=-1 --max-msgs-per-subject=-1 --max-age=1y --storage file --retention limits --max-msg-size=-1 --discard old --dupe-window="5s" --replicas 3

nats-list-streams: ## lists existing nats streams.
	@nats str ls

nats-describe-stream: ## describes a stream.
	@nats str info ORDERS

nats-backup-stream: ## copies a stream to another stream.
	@nats str cp ORDERS ARCHIVE --subjects "ORDERS_ARCHIVE.*" --max-age 2y

nats-edit-stream: ## edits the configuration of a stream.
	@nats str edit ARCHIVE --subjects "ARCHIVE.ORDERS.*"

nats-clear-stream: ## deletes all the data in a stream.
	@nats str purge ORDERS


# MAKE USAGE
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
