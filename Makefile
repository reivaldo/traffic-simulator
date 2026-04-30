# Makefile for Traffic Simulator

.PHONY: all start stop simulator-start simulator-stop k3s-setup k3s-status

all: start

start:
	docker-compose up -d

stop:
	docker-compose down

simulator-start:
	go run ./src/simulator/main.go --mode escalation

simulator-stop:
	pkill -f "go run ./src/simulator/main.go"

k3s-setup:
	bash k3s-cluster-setup.sh

k3s-status:
	sudo k3s kubectl get nodes
