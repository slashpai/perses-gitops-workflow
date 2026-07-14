.PHONY: validate render clean

PROJECT ?= monitoring
DATASOURCE ?= prometheus
OUTPUT_DIR ?= manifests/dashboards

validate:
	cd dashboards && go test ./...

render: validate
	cd dashboards && go run ./cmd/render \
		--project=$(PROJECT) \
		--datasource=$(DATASOURCE) \
		--output-dir=../$(OUTPUT_DIR)

clean:
	rm -rf manifests/dashboards/*.yaml
