#!/usr/bin/env bash
# Run every Fuzz* test in the given modules for a bounded time.
# One `go test -fuzz` call accepts a single function, so each match is run
# on its own. FUZZTIME limits each function (default 10s). Go test runs
# until that duration elapses.
#
# Usage: FUZZTIME=10s .github/ci-scripts/fuzz.sh ./modules/core ./cmd/wasm ./
set -euo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
cd "$root"

fuzztime=${FUZZTIME:-10s}
gocmd=${GOCMD:-go}

if [[ $# -eq 0 ]]; then
	echo "usage: $0 <module>..." >&2
	exit 1
fi

run_go() {
	if [[ -n "${GOWORK_ENV:-}" ]]; then
		env "${GOWORK_ENV}" "$gocmd" "$@"
	else
		"$gocmd" "$@"
	fi
}

for module in "$@"; do
	echo "→ Fuzzing ${module} (${fuzztime} each)"
	(
		cd "$module"
		# Skip vendor, VCS metadata, and nested modules so `./` does not
		# fuzz ./modules/core and ./cmd/wasm a second time.
		files=$(find . \( -name vendor -o -name .git \) -prune -o -type d ! -name . -exec test -f '{}/go.mod' \; -prune -o -name '*_test.go' -print)
		dirs=
		if [[ -n "$files" ]]; then
			matches=$(printf '%s\n' "$files" | xargs grep -l '^func Fuzz' || true)
			if [[ -n "$matches" ]]; then
				dirs=$(printf '%s\n' "$matches" | xargs -n1 dirname | sort -u)
			fi
		fi
		for dir in $dirs; do
			list_out=$(run_go test -list '^Fuzz' "$dir")
			names=$(printf '%s\n' "$list_out" | grep '^Fuzz' || true)
			for name in $names; do
				echo "→ ${name}"
				run_go test -run='^$' -fuzz="^${name}$" -fuzztime="$fuzztime" "$dir"
			done
		done
	)
done
