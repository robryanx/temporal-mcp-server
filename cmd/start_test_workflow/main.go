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
	name := flag.String("name", "test", "Name for the workflow input")
	failurePercent := flag.Float64("failure_percent", 0.1, "Percentage (0.0-1.0) of workflows that should fail")
	waitPercent := flag.Float64("wait_percent", 0.1, "Percentage (0.0-1.0) of workflows that should wait")
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

			// Determine the action based on percentages
			var workflowAction testworkflows.Action
			randVal := rand.Float64()
			if randVal < *failurePercent {
				workflowAction = testworkflows.Panic
			} else if randVal < *failurePercent+*waitPercent {
				workflowAction = testworkflows.Wait
			}

			input := testworkflows.SimpleWorkflowInput{
				Name:   *name,
				Action: workflowAction,
			}

			var we client.WorkflowRun
			switch *workflowName {
			case "test":
				we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, testworkflows.SimpleWorkflow, input)
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
