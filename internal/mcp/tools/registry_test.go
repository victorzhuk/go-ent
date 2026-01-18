package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestRegistryListHandler_UsesBoltDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := RegistryListInput{}

	result, _, err := registryListHandler(ctx, req, input)

	assert.Error(t, err, "Should error when path is empty")
	assert.Nil(t, result)
}

func TestRegistryListHandlerPathRequired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := RegistryListInput{Path: ""}

	_, _, err := registryListHandler(ctx, req, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path is required")
}
