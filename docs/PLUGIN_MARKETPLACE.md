# Plugin Marketplace Usage Guide

This guide explains how to use the go-ent plugin marketplace to discover, install, and manage plugins.

## Overview

The go-ent marketplace is a centralized repository for:
- Discovering community plugins
- Installing plugins with one command
- Managing plugin lifecycle (enable, disable, uninstall)
- Publishing your own plugins

**Default Marketplace URL**: `https://marketplace.go-ent.dev/api/v1`

## MCP Tools

The following MCP tools are available for marketplace operations:

| Tool | Purpose |
|------|---------|
| `plugin_list` | List installed plugins |
| `plugin_install` | Install plugin from marketplace or local path |
| `plugin_search` | Search plugins in marketplace |
| `plugin_info` | Get detailed plugin information |

## Listing Installed Plugins

Use `plugin_list` to see all installed plugins and their status.

### Example

```json
{
  "enabled": true
}
```

**Parameters**:
- `enabled` (optional, boolean): Filter by enabled status

**Response**:
```json
{
  "plugins": [
    {
      "name": "data-validation",
      "version": "1.2.3",
      "description": "Data validation patterns",
      "author": "Data Team",
      "enabled": true,
      "skills": 2,
      "agents": 1,
      "rules": 3
    },
    {
      "name": "go-perf",
      "version": "1.0.0",
      "description": "Performance patterns",
      "author": "Perf Team",
      "enabled": false,
      "skills": 1,
      "agents": 0,
      "rules": 0
    }
  ]
}
```

**Usage Examples**:

```bash
# List all installed plugins
plugin_list

# List only enabled plugins
plugin_list {"enabled": true}

# List only disabled plugins
plugin_list {"enabled": false}
```

## Searching Plugins

Use `plugin_search` to discover plugins in the marketplace.

### Example

```json
{
  "query": "validation",
  "category": "",
  "author": "",
  "sort_by": "downloads",
  "limit": 10
}
```

**Parameters**:
- `query` (required, string): Search query (name, description, tags)
- `category` (optional, string): Filter by category
- `author` (optional, string): Filter by author
- `sort_by` (optional, string): Sort order (`downloads`, `rating`, `name`, `updated`)
- `limit` (optional, integer): Maximum results (default: 20, max: 100)

**Response**:
```json
{
  "plugins": [
    {
      "name": "data-validation",
      "version": "1.2.3",
      "description": "Comprehensive data validation patterns and tools",
      "author": "Data Team",
      "category": "utilities",
      "downloads": 15234,
      "rating": 4.8,
      "tags": ["validation", "data", "schema"],
      "skills_count": 2,
      "agents_count": 1,
      "rules_count": 3
    },
    {
      "name": "input-validation",
      "version": "1.0.0",
      "description": "Simple input validation helpers",
      "author": "Utils Team",
      "category": "utilities",
      "downloads": 5432,
      "rating": 4.2,
      "tags": ["validation", "input"],
      "skills_count": 1,
      "agents_count": 0,
      "rules_count": 0
    }
  ],
  "total": 42
}
```

**Usage Examples**:

```bash
# Search for validation plugins
plugin_search {"query": "validation"}

# Search with sorting
plugin_search {"query": "database", "sort_by": "downloads"}

# Filter by author
plugin_search {"query": "", "author": "Data Team"}

# Filter by category
plugin_search {"query": "", "category": "utilities"}

# Limit results
plugin_search {"query": "security", "limit": 5}

# Combined search
plugin_search {
  "query": "api",
  "category": "development",
  "sort_by": "rating",
  "limit": 10
}
```

## Viewing Plugin Details

Use `plugin_info` to get detailed information about a plugin.

### Example

```json
{
  "name": "data-validation"
}
```

**Parameters**:
- `name` (required, string): Plugin name in marketplace

**Response**:
```json
{
  "name": "data-validation",
  "version": "1.2.3",
  "description": "Comprehensive data validation patterns and tools for enterprise applications",
  "author": "Data Team",
  "category": "utilities",
  "homepage": "https://github.com/data-team/validation-plugin",
  "repository": "https://github.com/data-team/validation-plugin",
  "license": "MIT",
  "downloads": 15234,
  "rating": 4.8,
  "rating_count": 123,
  "tags": ["validation", "data", "schema", "enterprise"],
  "skills_count": 2,
  "agents_count": 1,
  "rules_count": 3,
  "last_updated": "2026-01-15T10:30:00Z",
  "min_version": "3.0.0",
  "dependencies": []
}
```

**Usage Examples**:

```bash
# Get plugin info
plugin_info {"name": "data-validation"}

# Check if plugin meets requirements
plugin_info {"name": "advanced-plugin"}
# Then verify min_version field
```

## Installing Plugins from Marketplace

Use `plugin_install` to install plugins from the marketplace.

### Example

```json
{
  "name": "data-validation",
  "version": "latest"
}
```

**Parameters**:
- `name` (required, string): Plugin name in marketplace
- `version` (optional, string): Version to install (default: `latest`)

**Response**:
```json
{
  "success": true,
  "plugin": {
    "name": "data-validation",
    "version": "1.2.3",
    "description": "Comprehensive data validation patterns",
    "author": "Data Team",
    "enabled": false,
    "skills": 2,
    "agents": 1,
    "rules": 3
  },
  "message": "Plugin 'data-validation@1.2.3' installed successfully"
}
```

**Usage Examples**:

```bash
# Install latest version
plugin_install {"name": "data-validation"}

# Install specific version
plugin_install {"name": "data-validation", "version": "1.2.3"}

# Install with validation
plugin_install {"name": "my-plugin"}
# Plugin manager validates manifest automatically
```

## Installing from Local Path

You can install plugins from a local file path or URL.

### Example

```json
{
  "name": "local",
  "source": "/path/to/my-plugin.zip"
}
```

**Parameters**:
- `name` (required, string): Must be `"local"`
- `source` (required, string): Path to local plugin archive

**Usage Examples**:

```bash
# Install from local file
plugin_install {
  "name": "local",
  "source": "/home/user/my-plugin-1.0.0.zip"
}

# Install from relative path
plugin_install {
  "name": "local",
  "source": "./plugins/my-plugin.zip"
}

# Install from URL (if marketplace supports it)
plugin_install {
  "name": "local",
  "source": "https://example.com/plugins/my-plugin.zip"
}
```

## Enabling and Disabling Plugins

### Enable Plugin

After installation, plugins are disabled by default. Enable them with:

```bash
plugin_enable {"name": "data-validation"}
```

**Parameters**:
- `name` (required, string): Plugin name

**Response**:
```json
{
  "success": true,
  "plugin": {
    "name": "data-validation",
    "enabled": true,
    "skills": 2,
    "agents": 1,
    "rules": 3
  }
}
```

### Disable Plugin

Temporarily disable a plugin without uninstalling:

```bash
plugin_disable {"name": "data-validation"}
```

**Parameters**:
- `name` (required, string): Plugin name

**Response**:
```json
{
  "success": true,
  "plugin": {
    "name": "data-validation",
    "enabled": false,
    "skills": 2,
    "agents": 1,
    "rules": 3
  }
}
```

**Usage Examples**:

```bash
# Enable plugin
plugin_enable {"name": "data-validation"}

# Disable plugin
plugin_disable {"name": "data-validation"}

# Verify status
plugin_list {"enabled": true}
```

## Uninstalling Plugins

Remove a plugin from your system entirely:

```bash
plugin_uninstall {"name": "data-validation"}
```

**Parameters**:
- `name` (required, string): Plugin name

**Response**:
```json
{
  "success": true,
  "message": "Plugin 'data-validation' uninstalled successfully"
}
```

**Behavior**:
- Unregisters all skills, agents, and rules
- Removes plugin files from disk
- Clears configuration

**Usage Examples**:

```bash
# Uninstall plugin
plugin_uninstall {"name": "data-validation"}

# Verify removal
plugin_list
# Plugin should not appear in list
```

## Complete Workflow Example

Here's a complete workflow for discovering and using a plugin:

### Step 1: Search for Plugins

```bash
plugin_search {
  "query": "validation",
  "sort_by": "downloads",
  "limit": 10
}
```

**Response**:
```json
{
  "plugins": [
    {
      "name": "data-validation",
      "version": "1.2.3",
      "description": "Comprehensive data validation patterns",
      "author": "Data Team",
      "downloads": 15234,
      "rating": 4.8
    }
  ],
  "total": 1
}
```

### Step 2: Get Plugin Details

```bash
plugin_info {"name": "data-validation"}
```

**Response**:
```json
{
  "name": "data-validation",
  "version": "1.2.3",
  "description": "Comprehensive data validation patterns",
  "skills_count": 2,
  "agents_count": 1,
  "rules_count": 3,
  "min_version": "3.0.0"
}
```

### Step 3: Install Plugin

```bash
plugin_install {"name": "data-validation"}
```

**Response**:
```json
{
  "success": true,
  "message": "Plugin 'data-validation@1.2.3' installed successfully"
}
```

### Step 4: Verify Installation

```bash
plugin_list
```

**Response**:
```json
{
  "plugins": [
    {
      "name": "data-validation",
      "version": "1.2.3",
      "enabled": false,
      "skills": 2,
      "agents": 1,
      "rules": 3
    }
  ]
}
```

### Step 5: Enable Plugin

```bash
plugin_enable {"name": "data-validation"}
```

### Step 6: Verify Enabled

```bash
plugin_list {"enabled": true}
```

**Response**:
```json
{
  "plugins": [
    {
      "name": "data-validation",
      "enabled": true,
      "skills": 2,
      "agents": 1,
      "rules": 3
    }
  ]
}
```

### Step 7: Use Plugin

The plugin's skills, agents, and rules are now available:
- Skills will auto-activate when relevant
- Agents appear in agent selection
- Rules enforce constraints automatically

## Advanced Usage

### Managing Multiple Plugins

```bash
# List all installed
plugin_list

# List enabled only
plugin_list {"enabled": true}

# List disabled only
plugin_list {"enabled": false}

# Disable all plugins
for plugin in $(plugin_list | jq -r '.plugins[].name'); do
  plugin_disable {"name": "$plugin"}
done

# Enable specific plugins
plugin_enable {"name": "data-validation"}
plugin_enable {"name": "go-perf"}
```

### Version Management

```bash
# Check current version
plugin_info {"name": "data-validation"}

# Install specific version
plugin_install {"name": "data-validation", "version": "1.2.0"}

# Reinstall latest version
plugin_uninstall {"name": "data-validation"}
plugin_install {"name": "data-validation"}
```

### Conflict Resolution

If two plugins provide resources with the same name:

```bash
# Check what's installed
plugin_list

# Disable conflicting plugin
plugin_disable {"name": "plugin-a"}

# Enable preferred plugin
plugin_enable {"name": "plugin-b"}
```

### Dependency Management

When a plugin has `min_version` requirement:

```bash
# Check plugin requirements
plugin_info {"name": "advanced-plugin"}

# Output shows:
# "min_version": "3.0.0"

# Verify go-ent version meets requirement
go-ent --version

# If version is lower, upgrade go-ent first
# Then install plugin
plugin_install {"name": "advanced-plugin"}
```

## Troubleshooting

### Plugin Not Found

**Symptom**: `plugin_search` or `plugin_install` returns empty results

**Solutions**:
1. Check spelling of plugin name
2. Try broader search query
3. Verify marketplace is accessible
4. Check network connection

### Installation Fails

**Symptom**: `plugin_install` returns error

**Common Errors**:

```
Error: download failed: status 404, body: plugin not found
```
**Solution**: Plugin name or version doesn't exist

```
Error: validate manifest: name cannot be empty
```
**Solution**: Plugin manifest is invalid

```
Error: download size exceeds maximum allowed size of 104857600 bytes
```
**Solution**: Plugin archive too large (max 100MB)

### Plugin Won't Enable

**Symptom**: `plugin_enable` returns error

**Common Causes**:
- Skill name conflicts with existing skills
- Agent name conflicts with existing agents
- Missing dependencies
- Invalid manifest

**Solutions**:
1. Check logs for specific error
2. Disable conflicting plugins
3. Verify `min_version` requirement
4. Reinstall plugin

### Skills Not Activating

**Symptom**: Plugin enabled but skills don't activate

**Solutions**:
1. Verify plugin is enabled: `plugin_list {"enabled": true}`
2. Check skill description includes activation triggers
3. Restart go-ent to reload skills
4. Check logs for loading errors

### Marketplace Unavailable

**Symptom**: All marketplace operations fail

**Solutions**:
1. Check network connection
2. Verify marketplace URL is correct
3. Use local plugin installation as fallback
4. Check marketplace status page

## Best Practices

### Discovery

1. **Use specific search terms**: Query for exact functionality needed
2. **Check ratings and downloads**: Popular plugins are usually well-maintained
3. **Read descriptions**: Ensure plugin matches requirements
4. **Check dependencies**: Verify `min_version` compatibility

### Installation

1. **Start with latest version**: Unless specific version is required
2. **Install one at a time**: Test each plugin before adding more
3. **Verify after install**: Check plugin works as expected
4. **Enable only what's needed**: Keep enabled plugins minimal

### Management

1. **Regular updates**: Periodically check for new versions
2. **Disable unused plugins**: Reduce clutter and potential conflicts
3. **Monitor performance**: Some plugins may affect performance
4. **Uninstall obsolete**: Remove plugins no longer needed

### Development

1. **Test locally**: Before publishing, test installation from local path
2. **Validate manifest**: Use `plugin validate` before publishing
3. **Document clearly**: Provide good descriptions and examples
4. **Version carefully**: Follow semantic versioning

## CLI Commands

While MCP tools are recommended for interactive use, CLI commands are also available:

```bash
# List plugins
go-ent plugin list

# Search marketplace
go-ent plugin search "validation"

# Get plugin info
go-ent plugin info data-validation

# Install plugin
go-ent plugin install data-validation

# Install specific version
go-ent plugin install data-validation@1.2.3

# Install from local
go-ent plugin install local ./my-plugin.zip

# Enable plugin
go-ent plugin enable data-validation

# Disable plugin
go-ent plugin disable data-validation

# Uninstall plugin
go-ent plugin uninstall data-validation

# Validate manifest
go-ent plugin validate plugin.yaml
```

## Marketplace API

The marketplace REST API provides programmatic access:

### Search Plugins

```
GET /api/v1/plugins/search?q={query}&category={category}&author={author}&sort_by={sort_by}&limit={limit}
```

**Response**:
```json
{
  "plugins": [...],
  "total": 42
}
```

### Get Plugin Details

```
GET /api/v1/plugins/{name}
```

**Response**:
```json
{
  "name": "data-validation",
  "version": "1.2.3",
  ...
}
```

### Download Plugin

```
GET /api/v1/plugins/{name}/versions/{version}/download
```

**Response**: Binary plugin archive (zip)

## Configuration

### Marketplace URL

Default marketplace URL is `https://marketplace.go-ent.dev/api/v1`.

To use a custom marketplace:

```go
import "github.com/victorzhuk/go-ent/internal/marketplace"

client := marketplace.NewClientWithURL("https://custom.marketplace.com/api/v1")
```

### Plugin Directory

Default plugin directory is `./plugins` relative to go-ent working directory.

To use a custom directory:

```go
import "github.com/victorzhuk/go-ent/internal/plugin"

manager := plugin.NewManager("/custom/plugins/path", registry, marketplace, logger)
```

### Download Limits

Maximum download size is 100MB (100 * 1024 * 1024 bytes).

To change limit:

```go
const MaxDownloadSize = 200 * 1024 * 1024 // 200MB
```

## Security Considerations

### Validate Plugins

Before installing from marketplace:

1. Check plugin author and reputation
2. Review plugin description and capabilities
3. Check download count and rating
4. Read source code if available

### Local Plugins

When installing from local paths:

1. Verify source of plugin archive
2. Validate manifest: `plugin validate plugin.yaml`
3. Review plugin contents before enabling
4. Monitor for suspicious behavior

### Marketplace Trust

The marketplace does:
- Validate all manifests before publishing
- Scan for known vulnerabilities
- Review plugins manually if flagged

The marketplace does not:
- Guarantee plugin quality
- Verify all code paths
- Test all scenarios

## See Also

- **Plugin Development Guide**: `PLUGIN_DEVELOPMENT.md`
- **Manifest Reference**: `PLUGIN_MANIFEST.md`
- **Development Guide**: `DEVELOPMENT.md`
- **Marketplace Client Code**: `internal/marketplace/client.go`
- **Plugin Manager Code**: `internal/plugin/manager.go`
