#!/bin/bash

set -e

# Start RabbitMQ
/usr/local/bin/docker-entrypoint.sh rabbitmq-server -detached

# Join cluster
rabbitmqctl stop_app
rabbitmqctl join_cluster rabbit@rmq1 --ram

# Re-attach to rabbitmq server
rabbitmqctl stop
sleep 2s
rabbitmq-server
