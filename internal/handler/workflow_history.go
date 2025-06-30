package handler

import (
	"context"
	"fmt"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
)

type WorkflowHistoryArgs struct {
	WorkflowID string `json:"workflow_id" jsonschema:"required,description=The ID of the workflow to retrieve"`
	RunID      string `json:"run_id,omitempty" jsonschema:"description=Optional run ID of the workflow"`
}

type HistoryResponse struct {
	WorkflowID   string  `json:"workflow_id"`
	RunID        string  `json:"run_id"`
	Summary      string  `json:"summary"`
	Instructions string  `json:"instructionsForReadingEvents"`
	Events       []Event `json:"events"`
}

func GetWorkflowHistoryHandler(ctx context.Context, temporalClient client.Client, args WorkflowHistoryArgs) (HistoryResponse, error) {
	runID := args.RunID
	iter := temporalClient.GetWorkflowHistory(ctx, args.WorkflowID, runID, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)

	var rawEvents []*history.HistoryEvent
	for iter.HasNext() {
		evt, err := iter.Next()
		if err != nil {
			return HistoryResponse{}, fmt.Errorf("failed reading history: %w", err)
		}
		rawEvents = append(rawEvents, evt)
	}

	formattedList := make([]Event, 0, len(rawEvents))
	for _, e := range rawEvents {
		if formatted, keep := FormatEvent(e, false); keep {
			formattedList = append(formattedList, formatted)
		}
	}

	finalRunID := runID
	if len(rawEvents) > 0 && rawEvents[0].GetWorkflowExecutionStartedEventAttributes() != nil {
		finalRunID = rawEvents[0].GetWorkflowExecutionStartedEventAttributes().GetOriginalExecutionRunId()
	}

	return HistoryResponse{
		WorkflowID: args.WorkflowID,
		RunID:      finalRunID,
		Summary:    fmt.Sprintf("Workflow has %d events. We are examining %d events.", len(rawEvents), len(formattedList)),
		Events:     formattedList,
	}, nil
}


