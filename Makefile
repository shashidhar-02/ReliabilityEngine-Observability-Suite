.PHONY: help init plan apply lint test build push deploy-base deploy-apps clean

help:
	@echo "Enterprise SRE Platform Makefile"
	@echo "================================"
	@echo "Infrastructure:"
	@echo "  make init         - Initialize Terraform backend"
	@echo "  make plan         - Run Terraform plan"
	@echo "  make apply        - Apply Terraform configuration"
	@echo ""
	@echo "CI/CD & Testing:"
	@echo "  make lint         - Run linters (tflint, golangci-lint, flake8)"
	@echo "  make test         - Run unit tests for all microservices"
	@echo "  make build        - Build Docker images"
	@echo "  make scan         - Run security scan via Trivy"
	@echo ""
	@echo "Deployment:"
	@echo "  make deploy-base  - Deploy Kubernetes core manifestations"
	@echo "  make deploy-apps  - Deploy microservices via Helm"

init:
	terraform -chdir=terraform/env/prod init

plan:
	terraform -chdir=terraform/env/prod plan

apply:
	terraform -chdir=terraform/env/prod apply -auto-approve

lint:
	cd terraform/env/prod && tflint
	cd src/orders-api && golangci-lint run
	cd src/payments-api && flake8 .

test:
	cd src/orders-api && go test ./...
	cd src/payments-api && pytest

build:
	docker build -t orders-api:latest ./src/orders-api
	docker build -t payments-api:latest ./src/payments-api

scan: build
	trivy image --severity CRITICAL,HIGH orders-api:latest
	trivy image --severity CRITICAL,HIGH payments-api:latest

deploy-base:
	kubectl apply -k k8s-manifests/overlays/prod

deploy-apps:
	helm upgrade --install orders-api ./src/orders-api/chart --namespace default --atomic
	helm upgrade --install payments-api ./src/payments-api/chart --namespace default --atomic
