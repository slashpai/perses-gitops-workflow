.PHONY: validate-dashboards render-dashboards setup-prerequisites setup-argocd setup-metrics-usage check-metrics check-labels cleanup

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

setup-metrics-usage:
	kubectl apply -f deploy/metrics-usage/deployment.yaml
	kubectl -n perses-dev wait --for=condition=available deploy/metrics-usage --timeout=120s

check-metrics:
	@kubectl -n perses-dev port-forward svc/metrics-usage 18080:8080 >/dev/null 2>&1 & \
	PF_PID=$$!; \
	sleep 2; \
	PENDING=$$(curl -sf http://localhost:18080/api/v1/pending_usages 2>/dev/null); \
	kill $$PF_PID 2>/dev/null; \
	if [ -z "$$PENDING" ]; then \
		echo "FAIL: Could not reach metrics-usage API"; exit 1; \
	elif [ "$$PENDING" = "{}" ]; then \
		echo "OK: All dashboard metrics exist in Prometheus"; \
	else \
		echo "WARN: Dashboards reference metrics not found in Prometheus:"; \
		echo "$$PENDING"; \
	fi

check-labels:
	@kubectl -n perses-dev port-forward svc/metrics-usage 18080:8080 >/dev/null 2>&1 & \
	PF_PID=$$!; \
	sleep 2; \
	cd dashboards && go run ./cmd/checklabels \
		--manifests=../$(OUTPUT_DIR) \
		--metrics-usage-url=http://localhost:18080; \
	EXIT=$$?; \
	kill $$PF_PID 2>/dev/null; \
	exit $$EXIT

cleanup:
	bash ./scripts/cleanup.sh
