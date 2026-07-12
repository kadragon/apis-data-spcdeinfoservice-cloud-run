#!/bin/bash
set -euo pipefail

# Project root path detection
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "=== Running Harness & Convention Sweep Checks ==="

FAILED=0

# 1. Golden Principle 1: Service keys never in logs
# Verify no direct logging of serviceKey, authKey, DATAGOKR_SERVICEKEY, AUTH_API_KEY
# Ensure gin.Default() or gin.Logger() (which leak path parameters including service keys) are not used
echo "Checking for sensitive key leak in logs..."
if grep -rnE '(log|slog)\.(Print|Fatal|Panic|Info|Warn|Error|Debug)[a-zA-Z]*\(.*([sS]erviceKey|authKey|DATAGOKR_SERVICEKEY|AUTH_API_KEY).*\)' cmd/ internal/ \
     | perl -pe 's/\b(?:scrub[A-Za-z]*Key|redact[A-Za-z]+)\s*\([^)]*\)//g' \
     | grep -qE '([sS]erviceKey|authKey|DATAGOKR_SERVICEKEY|AUTH_API_KEY)'; then
  echo "❌ Error: Code contains direct logging of sensitive keys."
  grep -rnE '(log|slog)\.(Print|Fatal|Panic|Info|Warn|Error|Debug)[a-zA-Z]*\(.*([sS]erviceKey|authKey|DATAGOKR_SERVICEKEY|AUTH_API_KEY).*\)' cmd/ internal/ \
    | perl -pe 's/\b(?:scrub[A-Za-z]*Key|redact[A-Za-z]+)\s*\([^)]*\)//g' \
    | grep -E '([sS]erviceKey|authKey|DATAGOKR_SERVICEKEY|AUTH_API_KEY)' || true
  FAILED=1
fi

if grep -rn 'gin.Default()' cmd/ > /dev/null 2>&1; then
  echo "❌ Error: Do not use gin.Default() because it includes the default logger which logs raw query parameters (and leaks DATAGOKR_SERVICEKEY). Use gin.New() instead."
  grep -rn 'gin.Default()' cmd/ || true
  FAILED=1
fi

if grep -rn 'gin.Logger()' cmd/ > /dev/null 2>&1; then
  echo "❌ Error: Do not use gin.Logger() which logs raw query parameters. Use gin.New() and custom recovery middleware without default logging."
  grep -rn 'gin.Logger()' cmd/ || true
  FAILED=1
fi

# 2. Golden Principle 2: Auth middleware on all non-health routes
# Check that proxy.AuthMiddleware string appears in cmd/server/main.go (string-presence only; does not verify route scope)
echo "Checking for AuthMiddleware registration..."
if ! grep -q 'proxy.AuthMiddleware' cmd/server/main.go; then
  echo "❌ Error: AuthMiddleware registration missing in cmd/server/main.go"
  FAILED=1
fi

# 3. Golden Principle 3: New service API = new file (One ServiceSpec per file)
echo "Checking service specs convention..."
for file in internal/services/*.go; do
  filename=$(basename "$file")
  if [ "$filename" = "registry.go" ] || [[ "$filename" == *_test.go ]]; then
    continue
  fi
  count=$(grep -c "ServiceSpec{" "$file" || true)
  if [ "$count" -ne 1 ]; then
    echo "❌ Error: $file must contain exactly one ServiceSpec definition, found $count"
    FAILED=1
  fi
done

# 4. AGENTS.md Size check (target <= 100 lines, hard warn > 200)
echo "Checking AGENTS.md size limit..."
if [ -f "AGENTS.md" ]; then
  lines=$(wc -l < AGENTS.md | xargs)
  if [ "$lines" -gt 200 ]; then
    echo "❌ Error: AGENTS.md exceeds the hard size budget of 200 lines (current: $lines lines)"
    FAILED=1
  fi
else
  echo "❌ Error: Missing AGENTS.md"
  FAILED=1
fi

# 5. Check if required documentation exists
echo "Checking required documentation docs/*.md..."
docs=("architecture.md" "conventions.md" "delegation.md" "eval-criteria.md" "runbook.md" "workflows.md")
for doc in "${docs[@]}"; do
  if [ ! -f "docs/$doc" ]; then
    echo "❌ Error: Missing required documentation docs/$doc"
    FAILED=1
  fi
done

# 6. Go Formatting check
echo "Checking Go formatting..."
fmt_output=$(gofmt -l cmd/ internal/ || true)
if [ -n "$fmt_output" ]; then
  echo "❌ Error: The following files are not formatted with gofmt:"
  echo "$fmt_output"
  FAILED=1
fi

# 7. Check if tests pass and compile
echo "Running project tests..."
if ! go test ./... > /dev/null 2>&1; then
  echo "❌ Error: Tests are failing or code does not compile."
  FAILED=1
fi

if [ $FAILED -ne 0 ]; then
  echo "=== Sweep Failed! ==="
  exit 1
else
  echo "✅ Sweep passed successfully!"
  exit 0
fi
