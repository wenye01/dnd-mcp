// Package importer provides map import functionality
package importer

import (
	"github.com/dnd-mcp/server/internal/importer/format"
	"github.com/dnd-mcp/server/internal/importer/parser"
)

func init() {
	// Set the NDJSON parser factory function
	getNDJSONParser = func() ndjsonParser {
		return &ndjsonParserBridge{parser: parser.NewNDJSONParser()}
	}
}

// ndjsonParserBridge bridges the parser.NDJSONParser to the ndjsonParser interface
type ndjsonParserBridge struct {
	parser *parser.NDJSONParser
}

func (b *ndjsonParserBridge) Open(modulePath string) error {
	return b.parser.Open(modulePath)
}

func (b *ndjsonParserBridge) Close() error {
	return b.parser.Close()
}

func (b *ndjsonParserBridge) ListScenes() ([]string, error) {
	return b.parser.ListScenes()
}

func (b *ndjsonParserBridge) GetScene(nameOrID string) (*format.ParseResult, error) {
	return b.parser.GetScene(nameOrID)
}

func (b *ndjsonParserBridge) GetAllScenes() ([]*format.ParseResult, error) {
	return b.parser.GetAllScenes()
}

func (b *ndjsonParserBridge) GetModuleInfo() (*format.ModuleInfo, error) {
	return b.parser.GetModuleInfo()
}
