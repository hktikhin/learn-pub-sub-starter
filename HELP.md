# Install bootdev cli
go install github.com/bootdotdev/bootdev@latest

# Rabbitmq
go get github.com/rabbitmq/amqp091-go

./rabbit.sh start

docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management