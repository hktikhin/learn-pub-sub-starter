#!/bin/bash

start_or_run () {
    docker inspect rabbitmq-stomp > /dev/null 2>&1

    if [ $? -eq 0 ]; then
        echo "Starting Peril RabbitMQ container..."
        docker start rabbitmq-stomp
    else
        echo "Peril RabbitMQ container not found, creating a new one..."
        docker run -d --name rabbitmq-stomp -p 61613:61613 -p 15672:15672 rabbitmq-stomp
    fi
}

case "$1" in
    start)
        start_or_run
        ;;
    stop)
        echo "Stopping Peril RabbitMQ container..."
        docker stop rabbitmq-stomp
        ;;
    logs)
        echo "Fetching logs for Peril RabbitMQ container..."
        docker logs -f rabbitmq-stomp
        ;;
    *)
        echo "Usage: $0 {start|stop|logs}"
        exit 1
esac
