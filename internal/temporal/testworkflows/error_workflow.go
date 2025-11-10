package testworkflows

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// ErrorWorkflow is a workflow that demonstrates multiple activities and conditional errors.
func ErrorWorkflow(ctx workflow.Context, shouldError bool) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ErrorWorkflow started", "shouldError", shouldError)

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 5,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result1, result2 string
	var err error

	// Activity 1
	err = workflow.ExecuteActivity(ctx, Activity1, "first").Get(ctx, &result1)
	if err != nil {
		logger.Error("Activity1 failed", "error", err)
		return "", err
	}
	logger.Info("Activity1 completed", "result", result1)

	// Activity 2 - conditionally errors
	if shouldError {
		err = workflow.ExecuteActivity(ctx, Activity2, "second", true).Get(ctx, &result2)
	} else {
		err = workflow.ExecuteActivity(ctx, Activity2, "second", false).Get(ctx, &result2)
	}

	if err != nil {
		logger.Error("Activity2 failed", "error", err)
		return "", err
	}
	logger.Info("Activity2 completed", "result", result2)

	// Activity 3
	var result3 string
	err = workflow.ExecuteActivity(ctx, Activity3, "third").Get(ctx, &result3)
	if err != nil {
		logger.Error("Activity3 failed", "error", err)
		return "", err
	}
	logger.Info("Activity3 completed", "result", result3)

	finalResult := fmt.Sprintf("%s | %s | %s", result1, result2, result3)
	logger.Info("ErrorWorkflow completed", "finalResult", finalResult)
	return finalResult, nil
}

// Activity1 is a simple activity.
func Activity1(ctx context.Context, input string) (string, error) {
	activity.GetLogger(ctx).Info("Activity1 called", "input", input)
	return "Output from Activity1: " + input, nil
}

// Activity2 is an activity that can optionally return an error.
func Activity2(ctx context.Context, input string, returnError bool) (string, error) {
	activity.GetLogger(ctx).Info("Activity2 called", "input", input, "returnError", returnError)
	if returnError {
		return "", fmt.Errorf("intentional error from Activity2 for input: %s", input)
	}
	return "Output from Activity2: " + input, nil
}

// Activity3 is another simple activity.
func Activity3(ctx context.Context, input string) (string, error) {
	activity.GetLogger(ctx).Info("Activity3 called", "input", input)
	return "Output from Activity3: " + input, nil
}
