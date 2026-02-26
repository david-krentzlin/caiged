#!/usr/bin/env bash

set -u
set -o pipefail

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
MODULE_DIR="${QUALITY_GATE_MODULE_DIR:-$REPO_ROOT/caiged}"

if [[ ! -f "$MODULE_DIR/go.mod" ]]; then
	echo "Could not find go.mod at $MODULE_DIR"
	echo "Set QUALITY_GATE_MODULE_DIR to the Go module directory if needed."
	exit 1
fi

if ! command -v go >/dev/null 2>&1; then
	echo "go is required but was not found on PATH"
	exit 1
fi

MODULE_GO_VERSION=$(awk '/^go / {print $2; exit}' "$MODULE_DIR/go.mod")
if [[ -n "$MODULE_GO_VERSION" ]]; then
	GO_TOOLCHAIN_DEFAULT="go${MODULE_GO_VERSION}.0"
else
	GO_TOOLCHAIN_DEFAULT="auto"
fi
GO_TOOLCHAIN="${QUALITY_GATE_GOTOOLCHAIN:-$GO_TOOLCHAIN_DEFAULT}"

ARTIFACT_ROOT="${QUALITY_GATE_ARTIFACT_ROOT:-$REPO_ROOT/.artifacts/quality-gates}"
RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)"
ARTIFACT_DIR="$ARTIFACT_ROOT/$RUN_ID"

COVERAGE_THRESHOLD="${COVERAGE_THRESHOLD:-40.0}"
COMPLEXITY_THRESHOLD="${COMPLEXITY_THRESHOLD:-25}"
COMPLEXITY_TOP="${COMPLEXITY_TOP:-20}"
GOLANGCI_LINT_TIMEOUT="${GOLANGCI_LINT_TIMEOUT:-5m}"
INTEGRATION_TEST_TIMEOUT="${INTEGRATION_TEST_TIMEOUT:-5m}"
GOLANGCI_LINT_VERSION="${GOLANGCI_LINT_VERSION:-latest}"
GOCYCLO_VERSION="${GOCYCLO_VERSION:-v0.6.0}"

DEFAULT_GATES=(lint golangci-lint complexity test integration race coverage)
GOCYCLO_TARGETS=("$MODULE_DIR/cmd" "$MODULE_DIR/internal")

passed_gates=0
failed_gates=0

mkdir -p "$ARTIFACT_DIR"
cd "$REPO_ROOT" || exit 1

go_in_module() {
	GOTOOLCHAIN="$GO_TOOLCHAIN" go -C "$MODULE_DIR" "$@"
}

relative_path() {
	local raw_path="$1"
	if [[ "$raw_path" == "$REPO_ROOT/"* ]]; then
		printf "%s" "${raw_path#$REPO_ROOT/}"
		return
	fi
	printf "%s" "$raw_path"
}

print_gate_result() {
	local status="$1"
	local gate_name="$2"
	local elapsed_seconds="$3"
	local detail="$4"
	printf "[%s] %s (%ss) %s\n" "$status" "$gate_name" "$elapsed_seconds" "$detail"
}

run_gate() {
	local gate_name="$1"
	shift

	local log_file="$ARTIFACT_DIR/${gate_name}.log"
	local started_at ended_at elapsed_seconds

	started_at=$(date +%s)
	if "$@" >"$log_file" 2>&1; then
		passed_gates=$((passed_gates + 1))
		ended_at=$(date +%s)
		elapsed_seconds=$((ended_at - started_at))
		print_gate_result "PASS" "$gate_name" "$elapsed_seconds" "log: $(relative_path "$log_file")"
		return
	fi

	failed_gates=$((failed_gates + 1))
	ended_at=$(date +%s)
	elapsed_seconds=$((ended_at - started_at))
	print_gate_result "FAIL" "$gate_name" "$elapsed_seconds" "log: $(relative_path "$log_file")"
}

gate_lint() {
	local unformatted
	unformatted=$(cd "$MODULE_DIR" && gofmt -l ./main.go ./cmd ./internal ./integration)
	if [[ -n "$unformatted" ]]; then
		echo "Go files are not formatted. Run 'gofmt -w ./caiged/main.go ./caiged/cmd ./caiged/internal ./caiged/integration'"
		printf "%s\n" "$unformatted"
		return 1
	fi
	go_in_module vet ./...
}

gate_golangci_lint() {
	if command -v golangci-lint >/dev/null 2>&1; then
		(cd "$MODULE_DIR" && golangci-lint run --timeout "$GOLANGCI_LINT_TIMEOUT" ./...)
		return
	fi

	echo "golangci-lint not found on PATH. Using 'go run' fallback."
	go_in_module run "github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}" run --timeout "$GOLANGCI_LINT_TIMEOUT" ./...
}

gate_test() {
	go_in_module test ./...
}

gate_integration() {
	if ! command -v docker >/dev/null 2>&1; then
		echo "docker is required for integration tests"
		return 1
	fi
	if ! docker info >/dev/null 2>&1; then
		echo "docker daemon is not reachable"
		return 1
	fi
	go_in_module test -tags integration ./integration/... -timeout "$INTEGRATION_TEST_TIMEOUT"
}

gate_race() {
	go_in_module test -race ./cmd/... ./internal/...
}

gocyclo_runner() {
	if command -v gocyclo >/dev/null 2>&1; then
		gocyclo "$@"
		return
	fi
	go_in_module run "github.com/fzipp/gocyclo/cmd/gocyclo@${GOCYCLO_VERSION}" "$@"
}

run_coverage_gate() {
	local gate_name="coverage"
	local log_file="$ARTIFACT_DIR/${gate_name}.log"
	local profile_file="$ARTIFACT_DIR/coverage.out"
	local summary_file="$ARTIFACT_DIR/coverage-summary.txt"
	local started_at ended_at elapsed_seconds total

	started_at=$(date +%s)
	total=""

	if go_in_module test ./... -coverprofile="$profile_file" >"$log_file" 2>&1 && go_in_module tool cover -func="$profile_file" >"$summary_file" 2>>"$log_file"; then
		total=$(awk '/^total:/ {gsub("%", "", $3); print $3}' "$summary_file")
		if [[ -n "$total" ]] && awk -v actual="$total" -v threshold="$COVERAGE_THRESHOLD" 'BEGIN { exit (actual + 0 >= threshold + 0) ? 0 : 1 }'; then
			passed_gates=$((passed_gates + 1))
			ended_at=$(date +%s)
			elapsed_seconds=$((ended_at - started_at))
			print_gate_result "PASS" "$gate_name" "$elapsed_seconds" "coverage ${total}% >= ${COVERAGE_THRESHOLD}% | summary: $(relative_path "$summary_file")"
			return
		fi
	fi

	failed_gates=$((failed_gates + 1))
	ended_at=$(date +%s)
	elapsed_seconds=$((ended_at - started_at))
	if [[ -z "$total" ]]; then
		print_gate_result "FAIL" "$gate_name" "$elapsed_seconds" "coverage summary unavailable | log: $(relative_path "$log_file")"
		return
	fi
	print_gate_result "FAIL" "$gate_name" "$elapsed_seconds" "coverage ${total}% < ${COVERAGE_THRESHOLD}% | summary: $(relative_path "$summary_file")"
}

run_complexity_gate() {
	local gate_name="complexity"
	local log_file="$ARTIFACT_DIR/${gate_name}.log"
	local top_file="$ARTIFACT_DIR/complexity-top.txt"
	local avg_file="$ARTIFACT_DIR/complexity-avg.txt"
	local violations_file="$ARTIFACT_DIR/complexity-violations.txt"
	local started_at ended_at elapsed_seconds avg max violation_status

	started_at=$(date +%s)
	avg=""
	max="0"
	violation_status=0

	if ! gocyclo_runner -top "$COMPLEXITY_TOP" "${GOCYCLO_TARGETS[@]}" >"$top_file" 2>"$log_file"; then
		failed_gates=$((failed_gates + 1))
		ended_at=$(date +%s)
		elapsed_seconds=$((ended_at - started_at))
		print_gate_result "FAIL" "$gate_name" "$elapsed_seconds" "complexity scan failed | log: $(relative_path "$log_file")"
		return
	fi

	if ! gocyclo_runner -avg-short "${GOCYCLO_TARGETS[@]}" >"$avg_file" 2>>"$log_file"; then
		failed_gates=$((failed_gates + 1))
		ended_at=$(date +%s)
		elapsed_seconds=$((ended_at - started_at))
		print_gate_result "FAIL" "$gate_name" "$elapsed_seconds" "complexity average scan failed | log: $(relative_path "$log_file")"
		return
	fi

	avg=$(awk 'NF {value=$1} END {print value}' "$avg_file")
	max=$(awk 'NR==1 {print $1}' "$top_file")
	if [[ -z "$avg" ]]; then
		avg="0"
	fi
	if [[ -z "$max" ]]; then
		max="0"
	fi

	if ! gocyclo_runner -over "$COMPLEXITY_THRESHOLD" "${GOCYCLO_TARGETS[@]}" >"$violations_file" 2>>"$log_file"; then
		violation_status=$?
	fi

	ended_at=$(date +%s)
	elapsed_seconds=$((ended_at - started_at))

	if [[ "$violation_status" -eq 0 ]]; then
		passed_gates=$((passed_gates + 1))
		print_gate_result "PASS" "$gate_name" "$elapsed_seconds" "avg ${avg} max ${max} threshold ${COMPLEXITY_THRESHOLD} | report: $(relative_path "$top_file")"
		return
	fi

	failed_gates=$((failed_gates + 1))
	if [[ "$violation_status" -eq 1 ]]; then
		print_gate_result "FAIL" "$gate_name" "$elapsed_seconds" "avg ${avg} max ${max} threshold ${COMPLEXITY_THRESHOLD} | violations: $(relative_path "$violations_file")"
		return
	fi
	print_gate_result "FAIL" "$gate_name" "$elapsed_seconds" "complexity check failed | log: $(relative_path "$log_file")"
}

run_selected_gate() {
	local gate_name="$1"
	case "$gate_name" in
	lint)
		run_gate "lint" gate_lint
		;;
	golangci-lint)
		run_gate "golangci-lint" gate_golangci_lint
		;;
	test)
		run_gate "test" gate_test
		;;
	integration)
		run_gate "integration" gate_integration
		;;
	race)
		run_gate "race" gate_race
		;;
	coverage)
		run_coverage_gate
		;;
	complexity)
		run_complexity_gate
		;;
	all)
		local default_gate
		for default_gate in "${DEFAULT_GATES[@]}"; do
			run_selected_gate "$default_gate"
		done
		;;
	*)
		echo "Unknown quality gate target: $gate_name"
		echo "Available targets: all lint golangci-lint complexity test integration race coverage"
		failed_gates=$((failed_gates + 1))
		;;
	esac
}

if [[ "${1:-}" == "list" ]]; then
	echo "all"
	echo "lint"
	echo "golangci-lint"
	echo "complexity"
	echo "test"
	echo "integration"
	echo "race"
	echo "coverage"
	exit 0
fi

requested_targets=("$@")
if [[ "${requested_targets[0]:-}" == "run" ]]; then
	requested_targets=("${requested_targets[@]:1}")
fi
if [[ "${#requested_targets[@]}" -eq 0 ]]; then
	requested_targets=("all")
fi

for target in "${requested_targets[@]}"; do
	run_selected_gate "$target"
done

if ((failed_gates > 0)); then
	printf "Quality gates FAILED | passed=%d failed=%d | artifacts: %s\n" "$passed_gates" "$failed_gates" "$(relative_path "$ARTIFACT_DIR")"
	exit 1
fi

printf "Quality gates PASSED | passed=%d failed=%d | artifacts: %s\n" "$passed_gates" "$failed_gates" "$(relative_path "$ARTIFACT_DIR")"
