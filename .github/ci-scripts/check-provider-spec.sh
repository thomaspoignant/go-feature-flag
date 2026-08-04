#!/usr/bin/env bash
#
# Guards the GO Feature Flag provider conformance specification against drift.
#
# The specification pins an evaluation engine version (GOFF-ENG-001) and asserts
# expected evaluation results against canonical fixtures (Appendix B). Both can
# silently fall out of date: a routine dependency bump moves the engine, and a
# fixture edit moves the expected results. Neither would fail any existing test.
#
# Usage: .github/ci-scripts/check-provider-spec.sh
set -euo pipefail

SPEC="website/src/pages/specification/openfeature-provider.md"
WASM_GOMOD="cmd/wasm/go.mod"
RELEASE_MANIFEST=".github/release-please/.release-please-manifest.json"
FIXTURE="openfeature/providers/python-provider/tests/mock_responses/config/valid-all-types.json"

failures=0

fail() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}

pass() {
  echo "ok: $*"
}

for f in "$SPEC" "$WASM_GOMOD" "$RELEASE_MANIFEST" "$FIXTURE"; do
  if [[ ! -f "$f" ]]; then
    echo "FAIL: required file not found: $f" >&2
    exit 1
  fi
done

# ---------------------------------------------------------------------------
# GOFF-ENG-001 — the engine version the specification pins must be the engine
# version the WASM module is actually built from.
# ---------------------------------------------------------------------------

spec_core="$(grep -oE 'modules/core v[0-9]+\.[0-9]+\.[0-9]+' "$SPEC" | head -1 | sed 's|modules/core v||')"
wasm_core="$(grep -oE 'go-feature-flag/modules/core v[0-9]+\.[0-9]+\.[0-9]+' "$WASM_GOMOD" | head -1 | sed 's|.*modules/core v||')"

if [[ -z "$spec_core" ]]; then
  fail "could not find a 'modules/core vX.Y.Z' version in $SPEC"
elif [[ -z "$wasm_core" ]]; then
  fail "could not find the modules/core dependency in $WASM_GOMOD"
elif [[ "$spec_core" != "$wasm_core" ]]; then
  fail "engine version mismatch: spec pins modules/core v$spec_core, but $WASM_GOMOD builds against v$wasm_core.
      Update the 'Evaluation engine' row and GOFF-ENG-001 in $SPEC, and bump the
      specification version — conformance is defined relative to a specific engine."
else
  pass "engine version consistent (modules/core v$spec_core)"
fi

# ---------------------------------------------------------------------------
# GOFF-ENG-001 — the WASM module version named in the specification must be the
# one actually released. Deliberately compared against the release manifest and
# not against any single provider's pin: a provider lagging the current release
# is a conformance finding for that provider, not a defect in the specification.
# ---------------------------------------------------------------------------

spec_wasm="$(grep -oE 'WASM `[0-9]+\.[0-9]+\.[0-9]+`' "$SPEC" | head -1 | tr -d '`' | sed 's|WASM ||')"
released_wasm="$(grep -oE '"cmd/wasm": "[0-9]+\.[0-9]+\.[0-9]+"' "$RELEASE_MANIFEST" \
  | head -1 | sed 's|.*": "||; s|"||')"

if [[ -z "$spec_wasm" ]]; then
  fail "could not find a 'WASM \`X.Y.Z\`' version in $SPEC"
elif [[ -z "$released_wasm" ]]; then
  fail "could not find the cmd/wasm entry in $RELEASE_MANIFEST"
elif [[ "$spec_wasm" != "$released_wasm" ]]; then
  fail "WASM version mismatch: spec names $spec_wasm, latest release is $released_wasm.
      A new WASM release can change observable evaluation behaviour — 0.2.4 added
      structured PARSE_ERROR guards where 0.2.3 trapped. Review §10 and the
      specification version before bumping this."
else
  pass "WASM module version matches the latest release ($spec_wasm)"
fi

# ---------------------------------------------------------------------------
# Appendix B — every flag the expected-results table names must exist in the
# canonical fixture. A renamed or deleted flag would leave the specification
# asserting results for something that no longer exists.
# ---------------------------------------------------------------------------

appendix_flags=(
  bool_targeting_match
  string_key
  double_key
  integer_key
  object_key
  disabled_bool
)

missing=()
for flag in "${appendix_flags[@]}"; do
  if ! python3 -c "
import json, sys
with open('$FIXTURE') as fh:
    sys.exit(0 if '$flag' in json.load(fh).get('flags', {}) else 1)
"; then
    missing+=("$flag")
  fi
done

if [[ ${#missing[@]} -gt 0 ]]; then
  fail "flags referenced by Appendix B are absent from $FIXTURE: ${missing[*]}"
else
  pass "Appendix B references ${#appendix_flags[@]} flags, all present in the canonical fixture"
fi

# ---------------------------------------------------------------------------
# Appendix B — string_key must keep trackEvents:false. The specification uses it
# as the worked example for GOFF-COLL-022, so flipping it would make the
# specification wrong rather than merely stale.
# ---------------------------------------------------------------------------

if python3 -c "
import json, sys
with open('$FIXTURE') as fh:
    flags = json.load(fh)['flags']
sys.exit(0 if flags.get('string_key', {}).get('trackEvents') is False else 1)
"; then
  pass "string_key still sets trackEvents:false (GOFF-COLL-022 example intact)"
else
  fail "string_key no longer sets trackEvents:false in $FIXTURE.
      Appendix B and GOFF-COLL-022 use it as the worked example for a
      non-trackable flag."
fi

echo
if [[ "$failures" -gt 0 ]]; then
  echo "$failures check(s) failed." >&2
  exit 1
fi
echo "All provider specification consistency checks passed."
