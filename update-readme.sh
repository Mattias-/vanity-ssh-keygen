#!/usr/bin/env bash
set -euo pipefail

README_FILE="README.md"
START_MARKER="<!-- vanity-ssh-keygen-usage:start -->"
END_MARKER="<!-- vanity-ssh-keygen-usage:end -->"

# Run the help command and capture the output
HELP_OUTPUT=$(OVERRIDE_DEFAULT_THREADS=8 go run ./cmd/vanity-ssh-keygen --help)

# Extract the parts of the README before and after the markers
BEFORE_BLOCK=$(sed "/$START_MARKER/,\$d" "$README_FILE")
AFTER_BLOCK=$(sed "1,/$END_MARKER/d" "$README_FILE")

# Output new README content
cat <<EOF >"$README_FILE"
$BEFORE_BLOCK

$START_MARKER
\`\`\`text
$HELP_OUTPUT
\`\`\`
$END_MARKER
$AFTER_BLOCK
EOF

echo "✅ README.md updated!"
