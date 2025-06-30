package handler

import (
	"context"
	"fmt"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

type FailedWorkflow struct {
	WorkflowID string  `json:"workflow_id"`
	RunID      string  `json:"run_id"`
	Error      string  `json:"error"`
	Summary    []Event `json:"summary"`
}

type FailedWorkflowsResponse struct {
	Workflows []FailedWorkflow `json:"workflows"`
}

func GetFailedWorkflowsHandler(ctx context.Context, temporalClient client.Client) (FailedWorkflowsResponse, error) {
	var failedWorkflows []FailedWorkflow

	// 1. List open workflow executions
	resp, err := temporalClient.ListOpenWorkflow(ctx, &workflowservice.ListOpenWorkflowExecutionsRequest{})
	if err != nil {
		return FailedWorkflowsResponse{}, fmt.Errorf("failed to list open workflows: %w", err)
	}

	for _, wf := range resp.Executions {
		iter := temporalClient.GetWorkflowHistory(ctx, wf.Execution.GetWorkflowId(), wf.Execution.GetRunId(), false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)

		var summaryEvents []Event
		var errorMsg string

		for iter.HasNext() {
			evt, err := iter.Next()
			if err != nil {
				// Continue to next workflow if history is not available
				break
			}

			if formatted, keep := FormatEvent(evt, true); keep {
				summaryEvents = append(summaryEvents, formatted)
				if errVal, ok := formatted.Details["error"]; ok {
					if errMsg, isString := errVal.(string); isString {
						errorMsg = errMsg
					}
				}
			}
		}
		if errorMsg != "" {
			failedWorkflows = append(failedWorkflows, FailedWorkflow{
				WorkflowID: wf.Execution.GetWorkflowId(),
				RunID:      wf.Execution.GetRunId(),
				Error:      errorMsg,
				Summary:    summaryEvents,
			})
		}
	}

	return FailedWorkflowsResponse{Workflows: failedWorkflows}, nil
}
