package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/cmd/go-ent/internal/tools"
	"github.com/victorzhuk/go-ent/cmd/go-ent/internal/version"
)

func New() *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "go-ent",
			Version: version.String(),
		},
		nil,
	)

	tools.Register(s)

	return s
}
