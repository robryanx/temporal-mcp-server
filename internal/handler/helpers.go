package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/history/v1"
)

type Event struct {
	EventID   int64                  `json:"event_id"`
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

func FormatEvent(e *history.HistoryEvent, summaryOnly bool) (Event, bool) {
	if e == nil {
		return Event{}, false
	}

	details := make(map[string]interface{})

	if e.Attributes != nil {
		switch attr := e.Attributes.(type) {
		case *history.HistoryEvent_WorkflowExecutionStartedEventAttributes:
			if attr.WorkflowExecutionStartedEventAttributes != nil {
				extractJSON(attr.WorkflowExecutionStartedEventAttributes.GetInput(), &details)
			}

		case *history.HistoryEvent_ActivityTaskScheduledEventAttributes:
			if attr.ActivityTaskScheduledEventAttributes != nil {
				details["activity_type"] = attr.ActivityTaskScheduledEventAttributes.ActivityType.GetName()
				extractJSON(attr.ActivityTaskScheduledEventAttributes.Input, &details)
			}

		case *history.HistoryEvent_WorkflowExecutionSignaledEventAttributes:
			if attr.WorkflowExecutionSignaledEventAttributes != nil {
				details["signal_name"] = attr.WorkflowExecutionSignaledEventAttributes.GetSignalName()
				extractJSON(attr.WorkflowExecutionSignaledEventAttributes.Input, &details)
			}
		case *history.HistoryEvent_WorkflowExecutionUpdateAcceptedEventAttributes:
			if attr.WorkflowExecutionUpdateAcceptedEventAttributes != nil && attr.WorkflowExecutionUpdateAcceptedEventAttributes.AcceptedRequest != nil && attr.WorkflowExecutionUpdateAcceptedEventAttributes.AcceptedRequest.GetInput() != nil {
				updateInput := attr.WorkflowExecutionUpdateAcceptedEventAttributes.AcceptedRequest.GetInput()
				if updateInput.GetName() != "" {
					details["update_name"] = updateInput.GetName()
				}

				extractJSON(&common.Payloads{Payloads: updateInput.Args.Payloads}, &details)
			}
		case *history.HistoryEvent_ActivityTaskCompletedEventAttributes:
			if attr.ActivityTaskCompletedEventAttributes != nil {
				extractJSON(attr.ActivityTaskCompletedEventAttributes.Result, &details)
			}

		case *history.HistoryEvent_ActivityTaskFailedEventAttributes:
			if attr.ActivityTaskFailedEventAttributes != nil && attr.ActivityTaskFailedEventAttributes.Failure != nil {
				details["error"] = attr.ActivityTaskFailedEventAttributes.Failure.Message
			}

		case *history.HistoryEvent_WorkflowExecutionCompletedEventAttributes:
			if attr.WorkflowExecutionCompletedEventAttributes != nil {
				extractJSON(attr.WorkflowExecutionCompletedEventAttributes.Result, &details)
			}

		case *history.HistoryEvent_WorkflowExecutionFailedEventAttributes:
			if attr.WorkflowExecutionFailedEventAttributes != nil && attr.WorkflowExecutionFailedEventAttributes.Failure != nil {
				details["error"] = attr.WorkflowExecutionFailedEventAttributes.Failure.Message
			}

		case *history.HistoryEvent_TimerStartedEventAttributes:
			if attr.TimerStartedEventAttributes != nil {
				details["timer_id"] = attr.TimerStartedEventAttributes.GetTimerId()
				details["timeout"] = attr.TimerStartedEventAttributes.GetStartToFireTimeout().String()
			}

		case *history.HistoryEvent_WorkflowTaskFailedEventAttributes:
			if attr.WorkflowTaskFailedEventAttributes != nil && attr.WorkflowTaskFailedEventAttributes.Failure != nil {
				details["error"] = attr.WorkflowTaskFailedEventAttributes.Failure.Message + "\n" + attr.WorkflowTaskFailedEventAttributes.Failure.StackTrace
			}
		default:
			return Event{}, false
		}
	}

	return Event{
		EventID:   e.GetEventId(),
		Type:      e.GetEventType().String(),
		Timestamp: e.GetEventTime().UTC().Format(time.RFC3339),
		Details:   details,
	}, true
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
