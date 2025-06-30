package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/robryanx/mcp-temporal-server/internal/config"
	"github.com/robryanx/mcp-temporal-server/internal/handler"
	"github.com/robryanx/mcp-temporal-server/internal/temporal"
)

//go:embed instructions.txt
var instructions []byte

func main() {
	cfg := config.Load()
	tClient, err := temporal.NewTemporalClient(cfg)
	if err != nil {
		panic(err)
	}

	s := server.NewMCPServer(
		"Temporal MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
		server.WithInstructions(string(instructions)),
	)

	// Define the workflow_history tool schema
	tool := mcp.NewTool("workflow_history",
		mcp.WithDescription("Retrieve a workflow history"),
		mcp.WithString("workflow_id", mcp.Required(), mcp.Description("The ID of the workflow to retrieve")),
		mcp.WithString("run_id", mcp.Description("Optional run ID of the workflow")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workflowID, err := req.RequireString("workflow_id")
		if err != nil {
			return mcp.NewToolResultError("Missing required workflow_id"), nil
		}
		runID := req.GetString("run_id", "")
		args := handler.WorkflowHistoryArgs{WorkflowID: workflowID, RunID: runID}
		history, err := handler.GetWorkflowHistoryHandler(ctx, tClient, args)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		history.Instructions = string(instructions)
		jsonData, err := json.Marshal(history)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal history"), nil
		}
		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// Define the failed_workflows tool schema
	failedWorkflowsTool := mcp.NewTool("failed_workflows",
		mcp.WithDescription("Retrieve a list of workflows that are currently failing"),
	)

	s.AddTool(failedWorkflowsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		failedWorkflows, err := handler.GetFailedWorkflowsHandler(ctx, tClient)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonData, err := json.Marshal(failedWorkflows)
		if err != nil {
			return mcp.NewToolResultError("Failed to marshal failed workflows"), nil
		}
		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// Add instructions as a resource
	resource := mcp.NewResource("file://instructions", "instructions", mcp.WithMIMEType("text/plain"))
	s.AddResource(resource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "file://instructions",
				MIMEType: "text/plain",
				Text:     string(instructions),
			},
		}, nil
	})

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
