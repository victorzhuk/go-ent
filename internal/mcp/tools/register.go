package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

func Register(s *mcp.Server) {
	registerInit(s)
	registerList(s)
	registerShow(s)
	registerCRUD(s)
	registerRegistry(s)
	registerWorkflow(s)
	registerLoop(s)
	registerGenerate(s)
	registerValidate(s)
	registerArchive(s)
	registerListArchetypes(s)
	registerGenerateComponent(s)
	registerGenerateFromSpec(s)
	registerAgentExecute(s)
}
