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
	kubectl apply -f hack/metrics-usage/configmap.yaml
	kubectl apply -f hack/metrics-usage/cronjob.yaml

check-metrics:
	@kubectl -n perses-dev delete job metrics-usage-audit --ignore-not-found
	kubectl apply -f hack/metrics-usage/configmap.yaml -f hack/metrics-usage/audit-job.yaml
	@kubectl -n perses-dev wait --for=condition=complete job/metrics-usage-audit --timeout=600s
	@kubectl -n perses-dev logs job/metrics-usage-audit -c audit

check-labels:
	@kubectl -n perses-dev delete job metrics-usage-audit --ignore-not-found
	kubectl apply -f hack/metrics-usage/configmap.yaml -f hack/metrics-usage/audit-job.yaml
	@POD=$$(kubectl -n perses-dev get pods -l job-name=metrics-usage-audit -o jsonpath='{.items[0].metadata.name}'); \
	kubectl -n perses-dev wait --for=condition=Ready pod/$$POD --timeout=300s; \
	kubectl -n perses-dev port-forward pod/$$POD 18080:8080 >/dev/null 2>&1 & \
	PF_PID=$$!; \
	sleep 5; \
	cd hack/checklabels && go run . \
		--manifests=../../$(OUTPUT_DIR) \
		--metrics-usage-url=http://localhost:18080; \
	EXIT=$$?; \
	kill $$PF_PID 2>/dev/null; \
	exit $$EXIT

cleanup:
	bash ./scripts/cleanup.sh
