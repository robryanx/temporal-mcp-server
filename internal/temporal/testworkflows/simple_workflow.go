package testworkflows

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// Action is an enum for actions that can be taken in the workflow.
type Action string

const (
	// Panic is an action that causes the workflow to panic.
	Panic Action = "panic"
	// Wait is an action that causes the workflow to wait indefinitely.
	Wait Action = "wait"
)

// SimpleWorkflowInput is the input to the SimpleWorkflow.
type SimpleWorkflowInput struct {
	Name   string
	Action Action
}

// TestWorkflow is a simple workflow for testing purposes.
func SimpleWorkflow(ctx workflow.Context, input SimpleWorkflowInput) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("TestWorkflow started", "name", input.Name)

	// Panic if the action is "panic"
	if input.Action == Panic {
		panic("SimpleWorkflow: intentional panic triggered by input action 'panic'")
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result string
	err := workflow.ExecuteActivity(ctx, TestActivity, input.Name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed", "error", err)
		return "", err
	}

	// Wait indefinitely if the action is "wait"
	if input.Action == Wait {
		logger.Info("SimpleWorkflow: waiting indefinitely as requested by input action 'wait'")
		_ = workflow.Await(ctx, func() bool { return false }) // This condition will never be true
		return "Workflow waited", nil                         // This line will not be reached
	}

	logger.Info("TestWorkflow completed", "result", result)
	return result, nil
}

// TestActivity is a simple activity for testing purposes.
func TestActivity(ctx context.Context, name string) (string, error) {
	activity.GetLogger(ctx).Info("TestActivity called", "name", name)
	return "Hello, " + name + "!", nil
}
