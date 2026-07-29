.PHONY: validate-dashboards render-dashboards setup-prerequisites setup-argocd cleanup

PROJECT ?= perses-dev
DATASOURCE ?= prometheus-datasource
OUTPUT_DIR ?= manifests/dashboards

validate-dashboards:
	cd dashboards && go test ./...

render-dashboards: validate-dashboards
	cd dashboards && go run ./cmd/render \
		--project=$(PROJECT) \
		--datasource=$(DATASOURCE) \
		--output-dir=../$(OUTPUT_DIR)

setup-prerequisites:
	bash ./scripts/setup-prerequisites.sh

setup-argocd:
	bash ./scripts/setup-argocd.sh

cleanup:
	bash ./scripts/cleanup.sh
