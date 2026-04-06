#!/usr/bin/env bash
set -euo pipefail

README_FILE="README.md"
START_MARKER="<!-- vanity-ssh-keygen-usage:start -->"
END_MARKER="<!-- vanity-ssh-keygen-usage:end -->"

# Run the help command and capture the output
HELP_OUTPUT=$(OVERRIDE_DEFAULT_THREADS=8 ./vanity-ssh-keygen --help)

# Extract the parts of the README before and after the markers
BEFORE_BLOCK=$(sed "/$START_MARKER/,\$d" "$README_FILE")
AFTER_BLOCK=$(sed "1,/$END_MARKER/d" "$README_FILE")

# Generate new README content in memory
NEW_CONTENT=$(cat <<EOF
$BEFORE_BLOCK

$START_MARKER
\`\`\`text
$HELP_OUTPUT
\`\`\`
$END_MARKER
$AFTER_BLOCK
EOF
)

# Only update if the content is different
if ! echo "$NEW_CONTENT" | diff -q "$README_FILE" - > /dev/null; then
    echo "$NEW_CONTENT" > "$README_FILE"
    echo "✅ README.md updated!"
else
    echo "✅ README.md is already up to date."
fi
