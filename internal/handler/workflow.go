package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
)

type WorkflowHistoryArgs struct {
	WorkflowID string `json:"workflow_id" jsonschema:"required,description=The ID of the workflow to retrieve"`
	RunID      string `json:"run_id,omitempty" jsonschema:"description=Optional run ID of the workflow"`
}

type Event struct {
	EventID   int64                  `json:"event_id"`
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type HistoryResponse struct {
	WorkflowID string  `json:"workflow_id"`
	RunID      string  `json:"run_id"`
	Summary    string  `json:"summary"`
	Events     []Event `json:"events"`
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

	formatted := make([]Event, len(rawEvents))
	for i, e := range rawEvents {
		formatted[i] = FormatEvent(e)
	}

	finalRunID := runID
	if len(rawEvents) > 0 && rawEvents[0].GetWorkflowExecutionStartedEventAttributes() != nil {
		finalRunID = rawEvents[0].GetWorkflowExecutionStartedEventAttributes().GetOriginalExecutionRunId()
	}

	return HistoryResponse{
		WorkflowID: args.WorkflowID,
		RunID:      finalRunID,
		Summary:    fmt.Sprintf("Workflow has %d events.", len(formatted)),
		Events:     formatted,
	}, nil
}

func FormatEvent(e *history.HistoryEvent) Event {
	details := make(map[string]interface{})

	switch attr := e.Attributes.(type) {
	case *history.HistoryEvent_WorkflowExecutionStartedEventAttributes:
		extractJSON(attr.WorkflowExecutionStartedEventAttributes.Input, &details)

	case *history.HistoryEvent_ActivityTaskScheduledEventAttributes:
		details["activity_type"] = attr.ActivityTaskScheduledEventAttributes.ActivityType.GetName()
		extractJSON(attr.ActivityTaskScheduledEventAttributes.Input, &details)

	case *history.HistoryEvent_WorkflowExecutionSignaledEventAttributes:
		details["signal_name"] = attr.WorkflowExecutionSignaledEventAttributes.GetSignalName()
		extractJSON(attr.WorkflowExecutionSignaledEventAttributes.Input, &details)
	case *history.HistoryEvent_WorkflowExecutionUpdateAcceptedEventAttributes:
		if attr.WorkflowExecutionUpdateAcceptedEventAttributes.AcceptedRequest.GetInput() != nil {
			updateInput := attr.WorkflowExecutionUpdateAcceptedEventAttributes.AcceptedRequest.GetInput()
			if updateInput.GetName() != "" {
				details["update_name"] = updateInput.GetName()
			}

			extractJSON(&common.Payloads{Payloads: updateInput.Args.Payloads}, &details)
		}
	case *history.HistoryEvent_ActivityTaskCompletedEventAttributes:
		extractJSON(attr.ActivityTaskCompletedEventAttributes.Result, &details)

	case *history.HistoryEvent_ActivityTaskFailedEventAttributes:
		if attr.ActivityTaskFailedEventAttributes.Failure != nil {
			details["error"] = attr.ActivityTaskFailedEventAttributes.Failure.Message
		}

	case *history.HistoryEvent_WorkflowExecutionCompletedEventAttributes:
		extractJSON(attr.WorkflowExecutionCompletedEventAttributes.Result, &details)

	case *history.HistoryEvent_WorkflowExecutionFailedEventAttributes:
		if attr.WorkflowExecutionFailedEventAttributes.Failure != nil {
			details["error"] = attr.WorkflowExecutionFailedEventAttributes.Failure.Message
		}

	case *history.HistoryEvent_TimerStartedEventAttributes:
		details["timer_id"] = attr.TimerStartedEventAttributes.GetTimerId()
		details["timeout"] = attr.TimerStartedEventAttributes.GetStartToFireTimeout().String()
	}

	return Event{
		EventID:   e.GetEventId(),
		Type:      e.GetEventType().String(),
		Timestamp: e.GetEventTime().UTC().Format(time.RFC3339),
		Details:   details,
	}
}

func extractJSON(payload *common.Payloads, out *map[string]interface{}) {
	if payload == nil {
		return
	}

	data := payload.GetPayloads()
	if len(data) == 0 {
		return
	}

	for i, p := range data {
		var decoded interface{}
		if err := json.Unmarshal(p.GetData(), &decoded); err != nil {
			(*out)[fmt.Sprintf("input_part_%d", i)] = string(p.GetData())
		} else {
			(*out)[fmt.Sprintf("input_part_%d", i)] = decoded
		}
	}
}
