#!/bin/sh
# go-ent MCP server smart launcher
# Environment: GO_ENT_DEV=1|0, GO_ENT_VERBOSE=1

set -e

MODULE="github.com/victorzhuk/go-ent"
CMD_PATH="cmd/go-ent"
VERBOSE="${GO_ENT_VERBOSE:-0}"

log() { [ "$VERBOSE" = "1" ] && printf "[go-ent] %s\n" "$*" >&2 || true; }
die() { printf "[go-ent] ERROR: %s\n" "$*" >&2; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

find_project_root() {
    # Script is in scripts/, project root is 3 levels up
    # scripts/ -> go-ent/ -> plugins/ -> project root
    candidate="$(cd "$SCRIPT_DIR/../../.." 2>/dev/null && pwd)" || return 1
    [ -f "$candidate/go.mod" ] && echo "$candidate" && return 0
    return 1
}

PROJECT_ROOT="$(find_project_root)" || PROJECT_ROOT=""

detect_mode() {
    case "${GO_ENT_DEV:-}" in
        1|true|yes) echo "local"; return ;;
        0|false|no) echo "external"; return ;;
    esac

    if [ -n "$PROJECT_ROOT" ] && [ -f "$PROJECT_ROOT/go.mod" ]; then
        module_name="$(grep "^module " "$PROJECT_ROOT/go.mod" 2>/dev/null | awk '{print $2}')"
        [ "$module_name" = "$MODULE" ] && echo "local" && return
    fi
    echo "external"
}

run_local() {
    log "Running: go run ./$CMD_PATH"
    cd "$PROJECT_ROOT"
    exec go run "./$CMD_PATH" "$@"
}

main() {
    mode="$(detect_mode)"
    log "Mode: $mode"

    case "$mode" in
        local)
            [ -z "$PROJECT_ROOT" ] && die "Local mode but project root not found"
            run_local "$@"
            ;;
        external)
            if command -v go-ent >/dev/null 2>&1; then
                log "Using installed binary"
                exec go-ent "$@"
            elif command -v go >/dev/null 2>&1; then
                log "Remote fallback: go run $MODULE/$CMD_PATH@latest"
                exec go run "$MODULE/$CMD_PATH@latest" "$@"
            else
                die "No go-ent binary and 'go' not in PATH"
            fi
            ;;
    esac
}

main "$@"
