.PHONY: validate render setup-prerequisites cleanup

PROJECT ?= perses-dev
DATASOURCE ?= prometheus-datasource
OUTPUT_DIR ?= manifests/dashboards

validate:
	cd dashboards && go test ./...

render: validate
	cd dashboards && go run ./cmd/render \
		--project=$(PROJECT) \
		--datasource=$(DATASOURCE) \
		--output-dir=../$(OUTPUT_DIR)

setup-prerequisites:
	bash ./scripts/setup-prerequisites.sh

cleanup:
	bash ./scripts/cleanup.sh
