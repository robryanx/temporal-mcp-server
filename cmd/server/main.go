package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/robryanx/mcp-temporal-server/internal/config"
	"github.com/robryanx/mcp-temporal-server/internal/handler"
	"github.com/robryanx/mcp-temporal-server/internal/temporal"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

func main() {
	ctx := context.Background()

	done := make(chan struct{})

	cfg := config.Load()
	tClient, err := temporal.NewTemporalClient(cfg)
	if err != nil {
		panic(err)
	}

	server := mcp_golang.NewServer(stdio.NewStdioServerTransport())
	err = server.RegisterTool("workflow_history", "Retrieve a workflow history", func(arguments handler.WorkflowHistoryArgs) (*mcp_golang.ToolResponse, error) {
		history, err := handler.GetWorkflowHistoryHandler(ctx, tClient, arguments)
		if err != nil {
			return nil, err
		}

		jsonData, err := json.Marshal(history)
		if err != nil {
			return nil, err
		}

		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(jsonData))), nil
	})
	if err != nil {
		panic(err)
	}

	err = server.RegisterResource("test://resource", "guide", "Guide for understanding workflow histories", "text/plain", func() (*mcp_golang.ResourceResponse, error) {
		contents, err := os.ReadFile("instructions.txt")
		if err != nil {
			panic(err)
		}

		return mcp_golang.NewResourceResponse(mcp_golang.NewTextEmbeddedResource("test://resource", string(contents), "text/plain")), nil
	})

	err = server.Serve()
	if err != nil {
		panic(err)
	}

	<-done
}
