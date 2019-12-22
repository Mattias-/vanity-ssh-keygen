#!/bin/bash
set -euo pipefail

test_pubkey() {
    local filename="$1"
    # Only look in the second word of file
    cut -d' ' -f 2 "$filename" | grep -i "$find_me" >/dev/null
}

main() {
    local find_me="$1"
    local keyfile="./$RANDOM"
    while true; do
        ssh-keygen -q -f "$keyfile" -N '' -C '' -t ed25519 <<<'y' >/dev/null
        if test_pubkey "${keyfile}.pub"; then
            break
        fi
    done

    mv "$keyfile" "./${find_me}"
    mv "${keyfile}.pub" "./${find_me}.pub"

    echo "Found pubkey:"
    cat "./${find_me}.pub"
}

main "$@"
