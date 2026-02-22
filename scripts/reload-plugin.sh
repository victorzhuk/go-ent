#!/bin/sh
# Reload go-ent plugin in Claude Code
# Run this after making plugin changes (agents, skills, commands, or MCP server)

set -e

echo "🔧 Reloading go-ent plugin..."

# Clear plugin cache
if [ -d "$HOME/.claude/plugins/cache/go-ent" ]; then
    echo "  Removing cached plugin..."
    rm -rf "$HOME/.claude/plugins/cache/go-ent"
    echo "  ✓ Cache cleared"
else
    echo "  No cache found (this is fine)"
fi

# Rebuild MCP server binary
echo "  Rebuilding MCP server..."
cd "$(dirname "$0")/.."
make build > /dev/null 2>&1
echo "  ✓ MCP server rebuilt"

echo ""
echo "✅ Plugin ready for reload"
echo ""
echo "Next steps:"
echo "  1. Restart Claude Code"
echo "  2. Verify plugin loads (check status bar)"
echo "  3. Test MCP tools are available"
echo ""
