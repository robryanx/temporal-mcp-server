// Package testworkflows contains all test workflow and activity implementations.
package testworkflows

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// TestWorkflow is a simple workflow for testing purposes.
func TestWorkflow(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("TestWorkflow started", "name", name)

	// Panic if the name is "panic"
	if name == "panic" {
		panic("TestWorkflow: intentional panic triggered by input name 'panic'")
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result string
	err := workflow.ExecuteActivity(ctx, TestActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed", "error", err)
		return "", err
	}

	logger.Info("TestWorkflow completed", "result", result)
	return result, nil
}

// TestActivity is a simple activity for testing purposes.
func TestActivity(ctx context.Context, name string) (string, error) {
	activity.GetLogger(ctx).Info("TestActivity called", "name", name)
	return "Hello, " + name + "!", nil
}
