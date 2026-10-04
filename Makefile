ENV ?= prod
PYTHON ?= python3
IMAGE_TAG ?= latest

.PHONY: help setup init plan apply fmt format-check lint test validate helm-lint build scan deploy-base deploy-apps

help:
	@echo "ReliabilityEngine: setup, fmt, lint, test, validate, build, scan"
	@echo "Infrastructure: init, plan, apply (ENV=prod or ENV=staging)"
	@echo "Deployment: deploy-base, deploy-apps (ECR_REGISTRY and IMAGE_TAG required)"

setup:
	$(PYTHON) -m pip install --only-binary=:all: --require-hashes -r src/payments-api/requirements-dev.lock

init:
	terraform -chdir=terraform/env/$(ENV) init

plan:
	terraform -chdir=terraform/env/$(ENV) plan

apply:
	terraform -chdir=terraform/env/$(ENV) apply

fmt:
	gofmt -w src/orders-api/*.go
	terraform fmt -recursive terraform/

format-check:
	test -z "$$(gofmt -l src/orders-api)"
	terraform fmt -check -recursive terraform/

lint:
	cd src/orders-api && go vet ./...
	cd src/payments-api && $(PYTHON) -m flake8 .

test:
	cd src/orders-api && go test -race -cover ./...
	cd src/payments-api && $(PYTHON) -m pytest

helm-lint:
	helm lint src/orders-api/chart src/payments-api/chart
	helm template orders-api src/orders-api/chart -f src/orders-api/chart/values-prod.yaml >/dev/null
	helm template payments-api src/payments-api/chart -f src/payments-api/chart/values-prod.yaml >/dev/null

validate: format-check helm-lint
	kubectl kustomize k8s-manifests/overlays/prod >/dev/null
	terraform -chdir=terraform/env/prod init -backend=false -input=false
	terraform -chdir=terraform/env/prod validate
	terraform -chdir=terraform/env/staging init -backend=false -input=false
	terraform -chdir=terraform/env/staging validate

build:
	docker build -t orders-api:$(IMAGE_TAG) ./src/orders-api
	docker build -t payments-api:$(IMAGE_TAG) ./src/payments-api

scan: build
	trivy image --exit-code 1 --severity CRITICAL,HIGH orders-api:$(IMAGE_TAG)
	trivy image --exit-code 1 --severity CRITICAL,HIGH payments-api:$(IMAGE_TAG)

deploy-base:
	kubectl apply -k k8s-manifests/overlays/prod

deploy-apps:
	@test -n "$(ECR_REGISTRY)" && test "$(IMAGE_TAG)" != latest || (echo "Set ECR_REGISTRY and an immutable IMAGE_TAG" && exit 1)
	helm upgrade --install payments-api ./src/payments-api/chart --namespace default -f src/payments-api/chart/values-prod.yaml --set-string image.repository=$(ECR_REGISTRY)/payments-api --set-string image.tag=$(IMAGE_TAG) --atomic --wait
	helm upgrade --install orders-api ./src/orders-api/chart --namespace default -f src/orders-api/chart/values-prod.yaml --set-string image.repository=$(ECR_REGISTRY)/orders-api --set-string image.tag=$(IMAGE_TAG) --atomic --wait
