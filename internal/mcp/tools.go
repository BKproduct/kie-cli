// Copyright 2026 boris-korol. Licensed under Apache-2.0. See LICENSE.

package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BKproduct/kie-cli/internal/cli"
	"github.com/BKproduct/kie-cli/internal/client"
	"github.com/BKproduct/kie-cli/internal/cliutil"
	"github.com/BKproduct/kie-cli/internal/config"
	"github.com/BKproduct/kie-cli/internal/mcp/cobratree"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all API operations as MCP tools.
func RegisterTools(s *server.MCPServer) {
	s.AddTool(
		mcplib.NewTool("chat_credit",
			mcplib.WithDescription("Get remaining account credits. Returns the CreditResponse."),
			mcplib.WithReadOnlyHintAnnotation(true),
			mcplib.WithDestructiveHintAnnotation(false),
			mcplib.WithOpenWorldHintAnnotation(true),
		),
		makeAPIHandler("GET", "/api/v1/chat/credit", false, []mcpParamBinding{}, []string{}),
	)
	s.AddTool(
		mcplib.NewTool("common_download-url",
			mcplib.WithDescription("Get temporary download URL for generated file. Required: url. Returns the new DownloadUrlResponse."),
			mcplib.WithString("url", mcplib.Required(), mcplib.Description("Url")),
			mcplib.WithDestructiveHintAnnotation(false),
			mcplib.WithOpenWorldHintAnnotation(true),
		),
		makeAPIHandler("POST", "/api/v1/common/download-url", false, []mcpParamBinding{{PublicName: "url", WireName: "url", Location: "body"}}, []string{}),
	)
	s.AddTool(
		mcplib.NewTool("kie-ai-jobs_create-task",
			mcplib.WithDescription("Create async generation task. Required: input, model. Returns the new CreateTaskResponse."),
			mcplib.WithString("input", mcplib.Required(), mcplib.Description("Input")),
			mcplib.WithString("model", mcplib.Required(), mcplib.Description("Market model id e.g. nano-banana-pro, sora-2-text-to-video")),
			mcplib.WithDestructiveHintAnnotation(false),
			mcplib.WithOpenWorldHintAnnotation(true),
		),
		makeAPIHandler("POST", "/api/v1/jobs/createTask", false, []mcpParamBinding{{PublicName: "input", WireName: "input", Location: "body"}, {PublicName: "model", WireName: "model", Location: "body"}}, []string{}),
	)
	s.AddTool(
		mcplib.NewTool("kie-ai-jobs_record-info",
			mcplib.WithDescription("Get task status and result. Required: taskId. Returns the RecordInfoResponse."),
			mcplib.WithString("taskId", mcplib.Required(), mcplib.Description("Task id")),
			mcplib.WithReadOnlyHintAnnotation(true),
			mcplib.WithDestructiveHintAnnotation(false),
			mcplib.WithOpenWorldHintAnnotation(true),
		),
		makeAPIHandler("GET", "/api/v1/jobs/recordInfo", false, []mcpParamBinding{{PublicName: "taskId", WireName: "taskId", Location: "query"}}, []string{}),
	)

	// Context tool — front-loaded domain knowledge for agents.
	// Call this first to understand the API taxonomy, query patterns, and capabilities.
	s.AddTool(
		mcplib.NewTool("context",
			mcplib.WithDescription("Get API domain context: resource taxonomy, auth requirements, query tips, and unique capabilities. Call this first."),
			mcplib.WithReadOnlyHintAnnotation(true),
			mcplib.WithDestructiveHintAnnotation(false),
		),
		handleContext,
	)

	// Runtime Cobra-tree mirror — exposes every user-facing command that is
	// not already covered by a typed endpoint or framework MCP tool.
	cobratree.RegisterAll(s, cli.RootCmd(), cobratree.SiblingCLIPath)
}

type mcpParamBinding struct {
	PublicName string
	WireName   string
	Location   string
}

// makeAPIHandler creates a generic MCP tool handler for an API endpoint.
func makeAPIHandler(method, pathTemplate string, binaryResponse bool, bindings []mcpParamBinding, positionalParams []string) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		c, err := newMCPClient()
		if err != nil {
			return mcplib.NewToolResultError(err.Error()), nil
		}

		// mcp-go v0.47+ made CallToolParams.Arguments an `any` to support
		// non-map payloads; GetArguments() returns the map[string]any shape
		// we rely on here (or an empty map when the payload is something else).
		args := req.GetArguments()

		// positionalParams mixes real URL path params with CLI positional
		// args that map to query params (e.g. `search <query>` -> ?query=);
		// the placeholder check below disambiguates them at runtime.
		path := pathTemplate
		knownArgs := make(map[string]bool, len(bindings))
		pathParams := make(map[string]bool, len(positionalParams))
		params := make(map[string]string)
		bodyArgs := make(map[string]any)
		var headers map[string]string
		if binaryResponse {
			headers = map[string]string{client.BinaryResponseHeader: "true"}
		}
		for _, binding := range bindings {
			knownArgs[binding.PublicName] = true
			v, ok := args[binding.PublicName]
			if !ok {
				continue
			}
			switch binding.Location {
			case "path":
				placeholder := "{" + binding.WireName + "}"
				pathParams[binding.PublicName] = true
				path = strings.Replace(path, placeholder, fmt.Sprintf("%v", v), 1)
			case "body":
				bodyArgs[binding.WireName] = v
			default:
				params[binding.WireName] = fmt.Sprintf("%v", v)
			}
		}
		for _, p := range positionalParams {
			placeholder := "{" + p + "}"
			if !strings.Contains(pathTemplate, placeholder) {
				continue
			}
			pathParams[p] = true
			if v, ok := args[p]; ok {
				path = strings.Replace(path, placeholder, fmt.Sprintf("%v", v), 1)
			}
		}

		for k, v := range args {
			if pathParams[k] || knownArgs[k] {
				continue
			}
			switch method {
			case "POST", "PUT", "PATCH":
				bodyArgs[k] = v
			default:
				params[k] = fmt.Sprintf("%v", v)
			}
		}

		var data json.RawMessage
		switch method {
		case "GET":
			if binaryResponse {
				data, err = c.GetWithHeaders(path, params, headers)
				break
			}
			data, err = c.Get(path, params)
		case "POST":
			if binaryResponse {
				data, _, err = c.PostWithParamsAndHeaders(path, params, bodyArgs, headers)
				break
			}
			data, _, err = c.PostWithParams(path, params, bodyArgs)
		case "PUT":
			if binaryResponse {
				data, _, err = c.PutWithParamsAndHeaders(path, params, bodyArgs, headers)
				break
			}
			data, _, err = c.PutWithParams(path, params, bodyArgs)
		case "PATCH":
			if binaryResponse {
				data, _, err = c.PatchWithParamsAndHeaders(path, params, bodyArgs, headers)
				break
			}
			data, _, err = c.PatchWithParams(path, params, bodyArgs)
		case "DELETE":
			if binaryResponse {
				data, _, err = c.DeleteWithParamsAndHeaders(path, params, headers)
				break
			}
			data, _, err = c.DeleteWithParams(path, params)
		default:
			return mcplib.NewToolResultError("unsupported method: " + method), nil
		}

		if err != nil {
			msg := err.Error()
			switch {
			case strings.Contains(msg, "HTTP 409"):
				return mcplib.NewToolResultText("already exists (no-op)"), nil
			case strings.Contains(msg, "HTTP 400") && cliutil.LooksLikeAuthError(msg):
				return mcplib.NewToolResultError("authentication error: " + cliutil.SanitizeErrorBody(msg) +
					"\nhint: the API rejected the request — this usually means auth is missing or invalid." +
					"\n      Set your API key: export KIE_BEARER_AUTH=<your-key>" +
					"\n      Get a key at: https://kie.ai/api-key" +
					"\n      Run 'kie-cli doctor' to check auth status."), nil
			case strings.Contains(msg, "HTTP 401"):
				return mcplib.NewToolResultError("authentication failed: " + cliutil.SanitizeErrorBody(msg) +
					"\nhint: check your token." +
					"\n      Set it with: export KIE_BEARER_AUTH=<your-key>" +
					"\n      Get a key at: https://kie.ai/api-key" +
					"\n      Run 'kie-cli doctor' to check auth status."), nil
			case strings.Contains(msg, "HTTP 403"):
				return mcplib.NewToolResultError("permission denied: " + cliutil.SanitizeErrorBody(msg) +
					"\nhint: your credentials are valid but lack access to this resource." +
					"\n      Set it with: export KIE_BEARER_AUTH=<your-key>" +
					"\n      Get a key at: https://kie.ai/api-key" +
					"\n      Run 'kie-cli doctor' to check auth status."), nil
			case strings.Contains(msg, "HTTP 404"):
				if method == "DELETE" {
					return mcplib.NewToolResultText("already deleted (no-op)"), nil
				}
				return mcplib.NewToolResultError("not found: " + msg), nil
			case strings.Contains(msg, "HTTP 429"):
				return mcplib.NewToolResultError("rate limited: " + msg), nil
			default:
				return mcplib.NewToolResultError(msg), nil
			}
		}

		// For GET responses, wrap bare arrays with count metadata
		if method == "GET" {
			trimmed := strings.TrimSpace(string(data))
			if len(trimmed) > 0 && trimmed[0] == '[' {
				var items []json.RawMessage
				if json.Unmarshal(data, &items) == nil {
					wrapped := map[string]any{
						"count": len(items),
						"items": items,
					}
					out, _ := json.Marshal(wrapped)
					return mcplib.NewToolResultText(string(out)), nil
				}
			}
		}
		if binaryResponse {
			out, _ := json.Marshal(map[string]any{
				"content_encoding": "base64",
				"data_base64":      base64.StdEncoding.EncodeToString(data),
				"byte_count":       len(data),
			})
			return mcplib.NewToolResultText(string(out)), nil
		}
		return mcplib.NewToolResultText(string(data)), nil
	}
}

func newMCPClient() (*client.Client, error) {
	home, _ := os.UserHomeDir()
	cfgPath := filepath.Join(home, ".config", "kie-cli", "config.toml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	c := client.New(cfg, 30*time.Second, 0)
	// Agents calling through MCP need fresh data every call. The on-disk
	// response cache survives across MCP server invocations, so a
	// DELETE/PATCH followed by a GET would otherwise return the
	// pre-mutation snapshot for up to the cache TTL. The interactive CLI
	// constructs its own client and is unaffected.
	c.NoCache = true
	return c, nil
}

func dbPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "kie-cli", "data.db")
}

// Note: MCP tools use their own dbPath() because they are in a separate package (main, not cli).
// The CLI's defaultDBPath() in the cli package uses the same canonical path.

func handleContext(_ context.Context, _ mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	ctx := map[string]any{
		"api":         "kie",
		"description": "Unified job API for KIE.ai market models (image/video generation). Bearer token from https://kie.ai/api-key — env...",
		"archetype":   "communication",
		"tool_count":  4,
		// tool_surface tells agents which surface a capability lives on.
		"tool_surface": "MCP exposes typed endpoint tools plus a runtime mirror of user-facing CLI commands. Endpoint tools keep typed schemas; command-mirror tools shell out to the companion kie-cli binary.",
		"auth": map[string]any{
			"type": "bearer_token",
			"env_vars": []map[string]any{
				{
					"name":        "KIE_BEARER_AUTH",
					"kind":        "per_call",
					"required":    true,
					"sensitive":   true,
					"description": "Set to your API credential.",
				},
			},
			"key_url": "https://kie.ai/api-key",
		},
		"resources": []map[string]any{
			{
				"name":        "chat",
				"description": "Manage chat",
				"endpoints":   []string{"credit"},
				"syncable":    true,
			},
			{
				"name":        "common",
				"description": "Manage common",
				"endpoints":   []string{"download-url"},
				"searchable":  true,
			},
			{
				"name":        "kie-ai-jobs",
				"description": "Manage kie ai jobs",
				"endpoints":   []string{"create-task", "record-info"},
				"searchable":  true,
			},
		},
		"query_tips": []string{
			"Pagination uses cursor-based paging. Pass after parameter for subsequent pages.",
			"Control page size with the limit parameter (default 100).",
		},
		"playbook": []map[string]string{
			{"topic": "Message search", "insight": "Use the search tool on synced data rather than paginating through message history. Message APIs often have aggressive rate limits."},
			{"topic": "Channel health", "insight": "When analyzing channel activity, use the channel-health command or sql aggregation on synced messages. Don't iterate individual messages via API."},
		},
	}
	data, _ := json.MarshalIndent(ctx, "", "  ")
	return mcplib.NewToolResultText(string(data)), nil
}

// RegisterNovelFeatureTools is kept as a compatibility no-op for older MCP
// mains. New generated mains call RegisterTools only; RegisterTools now
// includes the runtime Cobra-tree mirror.
func RegisterNovelFeatureTools(s *server.MCPServer) {
	_ = s
}
