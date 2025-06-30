package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/robryanx/mcp-temporal-server/internal/config"
	"github.com/robryanx/mcp-temporal-server/internal/temporal"
	"github.com/robryanx/mcp-temporal-server/internal/temporal/testworkflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	workflowName := flag.String("workflow", "test", "Name of the workflow to start (e.g. 'test')")
	count := flag.Int("count", 1, "Number of workflows to start")
	input := flag.String("input", "test", "Input string for the workflow")
	failures := flag.Float64("failures", 0.1, "Percentage (0.0-1.0) of workflows that should fail when input is 'random'")
	flag.Parse()

	cfg := config.Load()

	c, err := temporal.NewTemporalClient(cfg)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	w := worker.New(c, "test-task-queue", worker.Options{})

	// Register all test workflows here
	w.RegisterWorkflow(testworkflows.SimpleWorkflow)
	w.RegisterActivity(testworkflows.TestActivity)

	// Start workflows in a goroutine, then run the worker in the main goroutine (blocking)
	go func() {
		for i := 0; i < *count; i++ {
			workflowID := fmt.Sprintf("%s-workflow-%d-%d", *workflowName, time.Now().Unix(), i)
			workflowOptions := client.StartWorkflowOptions{
				ID:        workflowID,
				TaskQueue: "test-task-queue",
			}

			// Use the -failures flag to determine failure rate if input is "random"
			inputVal := *input
			if *input == "random" && rand.Float64() < *failures {
				inputVal = "panic"
			}

			var we client.WorkflowRun
			switch *workflowName {
			case "test":
				we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, testworkflows.SimpleWorkflow, inputVal)
			default:
				fmt.Printf("Unknown workflow: %s\n", *workflowName)
				continue
			}
			if err != nil {
				fmt.Printf("Failed to start workflow: %v\n", err)
				continue
			}

			fmt.Printf("Started workflow. WorkflowID: %s RunID: %s\n", we.GetID(), we.GetRunID())
		}
	}()

	// Run the worker in the main goroutine (blocking)
	err = w.Run(worker.InterruptCh())
	if err != nil {
		fmt.Println("Worker error:", err)
		os.Exit(1)
	}
}
