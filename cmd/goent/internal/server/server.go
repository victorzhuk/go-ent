package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/cmd/goent/internal/tools"
)

const Version = "2.0.0"

func New() *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "goent-spec",
			Version: Version,
		},
		nil,
	)

	tools.Register(s)

	return s
}
