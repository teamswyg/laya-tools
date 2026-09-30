package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/app"
	"io"
)

// stdio MCP uses newline-delimited JSON-RPC. stdout is protocol-only.
func serveMCP(a *app.App, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	enc := json.NewEncoder(out)
	for scanner.Scan() {
		var m struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      json.RawMessage `json:"id"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
			if err = enc.Encode(map[string]any{"jsonrpc": "2.0", "id": nil, "error": map[string]any{"code": -32700, "message": "parse error"}}); err != nil {
				return err
			}
			continue
		}
		if len(m.ID) == 0 {
			continue
		}
		response := map[string]any{"jsonrpc": "2.0", "id": m.ID}
		var result any
		switch m.Method {
		case "initialize":
			result = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]string{"name": "laya-tools", "version": version}}
		case "ping":
			result = map[string]any{}
		case "tools/list":
			result = map[string]any{"tools": []any{
				map[string]any{"name": "search_code", "description": "Retrieve bounded code excerpts from the configured local repository. English identifiers improve candidate recall.", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"query": map[string]string{"type": "string"}, "candidate_query": map[string]string{"type": "string"}, "lexical": map[string]string{"type": "boolean"}}, "required": []string{"query"}, "additionalProperties": false}},
				map[string]any{"name": "route_model", "description": "Recommend a configured Codex model for a NEW task; does not change the running conversation or execute anything.", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"query": map[string]string{"type": "string"}}, "required": []string{"query"}, "additionalProperties": false}}}}
		case "tools/call":
			var p struct {
				Name      string      `json:"name"`
				Arguments app.Request `json:"arguments"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				response["error"] = map[string]any{"code": -32602, "message": "invalid arguments"}
				break
			}
			switch p.Name {
			case "search_code":
				p.Arguments.Op = "search"
			case "route_model":
				p.Arguments.Op = "route"
			default:
				response["error"] = map[string]any{"code": -32602, "message": "unknown tool"}
			}
			if response["error"] == nil {
				r := a.Process(p.Arguments)
				b, err := json.Marshal(r)
				if err != nil {
					return err
				}
				result = map[string]any{"content": []any{map[string]string{"type": "text", "text": string(b)}}, "isError": r.Error != ""}
			}
		default:
			response["error"] = map[string]any{"code": -32601, "message": fmt.Sprintf("unsupported method %s", m.Method)}
		}
		if response["error"] == nil {
			response["result"] = result
		}
		if err := enc.Encode(response); err != nil {
			return err
		}
	}
	return scanner.Err()
}
