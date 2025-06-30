package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/failure/v1"
)

func TestFormatEvent_NilEvent(t *testing.T) {
	event, keep := FormatEvent(nil, false)
	assert.False(t, keep)
	assert.Equal(t, Event{}, event)
}

func TestFormatEvent_WorkflowExecutionStartedEvent(t *testing.T) {
	now := time.Now()
	event := &history.HistoryEvent{
		EventId:   1,
		EventTime: &now,
		EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
		Attributes: &history.HistoryEvent_WorkflowExecutionStartedEventAttributes{
			WorkflowExecutionStartedEventAttributes: &history.WorkflowExecutionStartedEventAttributes{
				Input: &common.Payloads{
					Payloads: []*common.Payload{
						{
							Data: []byte(`"test_input"`),
						},
					},
				},
				OriginalExecutionRunId: "run-123",
			},
		},
	}

	formattedEvent, keep := FormatEvent(event, false)
	assert.True(t, keep)
	assert.Equal(t, int64(1), formattedEvent.EventID)
	assert.Equal(t, "WorkflowExecutionStarted", formattedEvent.Type)
	assert.Equal(t, now.UTC().Format(time.RFC3339), formattedEvent.Timestamp)
	assert.Contains(t, formattedEvent.Details, "input_part_0")
	assert.Equal(t, "test_input", formattedEvent.Details["input_part_0"])
}

func TestFormatEvent_ActivityTaskScheduledEvent(t *testing.T) {
	now := time.Now()
	event := &history.HistoryEvent{
		EventId:   2,
		EventTime: &now,
		EventType: enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
		Attributes: &history.HistoryEvent_ActivityTaskScheduledEventAttributes{
			ActivityTaskScheduledEventAttributes: &history.ActivityTaskScheduledEventAttributes{
				ActivityType: &common.ActivityType{Name: "test_activity"},
				Input: &common.Payloads{
					Payloads: []*common.Payload{
						{
							Data: []byte(`"activity_input"`),
						},
					},
				},
			},
		},
	}

	formattedEvent, keep := FormatEvent(event, false)
	assert.True(t, keep)
	assert.Equal(t, int64(2), formattedEvent.EventID)
	assert.Equal(t, "ActivityTaskScheduled", formattedEvent.Type)
	assert.Equal(t, now.UTC().Format(time.RFC3339), formattedEvent.Timestamp)
	assert.Contains(t, formattedEvent.Details, "activity_type")
	assert.Equal(t, "test_activity", formattedEvent.Details["activity_type"])
	assert.Contains(t, formattedEvent.Details, "input_part_0")
	assert.Equal(t, "activity_input", formattedEvent.Details["input_part_0"])
}

func TestFormatEvent_ActivityTaskFailedEvent(t *testing.T) {
	now := time.Now()
	event := &history.HistoryEvent{
		EventId:   3,
		EventTime: &now,
		EventType: enums.EVENT_TYPE_ACTIVITY_TASK_FAILED,
		Attributes: &history.HistoryEvent_ActivityTaskFailedEventAttributes{
			ActivityTaskFailedEventAttributes: &history.ActivityTaskFailedEventAttributes{
				Failure: &failure.Failure{Message: "activity failed"},
			},
		},
	}

	formattedEvent, keep := FormatEvent(event, false)
	assert.True(t, keep)
	assert.Equal(t, int64(3), formattedEvent.EventID)
	assert.Equal(t, "ActivityTaskFailed", formattedEvent.Type)
	assert.Equal(t, now.UTC().Format(time.RFC3339), formattedEvent.Timestamp)
	assert.Contains(t, formattedEvent.Details, "error")
	assert.Equal(t, "activity failed", formattedEvent.Details["error"])
}

func TestFormatEvent_TimerStartedEvent(t *testing.T) {
	now := time.Now()
	d := time.Second * 10
	event := &history.HistoryEvent{
		EventId:   4,
		EventTime: &now,
		EventType: enums.EVENT_TYPE_TIMER_STARTED,
		Attributes: &history.HistoryEvent_TimerStartedEventAttributes{
			TimerStartedEventAttributes: &history.TimerStartedEventAttributes{
				TimerId:            "timer-1",
				StartToFireTimeout: &d,
			},
		},
	}

	formattedEvent, keep := FormatEvent(event, false)
	assert.True(t, keep)
	assert.Equal(t, int64(4), formattedEvent.EventID)
	assert.Equal(t, "TimerStarted", formattedEvent.Type)
	assert.Equal(t, now.UTC().Format(time.RFC3339), formattedEvent.Timestamp)
	assert.Contains(t, formattedEvent.Details, "timer_id")
	assert.Equal(t, "timer-1", formattedEvent.Details["timer_id"])
	assert.Contains(t, formattedEvent.Details, "timeout")
	assert.Equal(t, "10s", formattedEvent.Details["timeout"])
}

func TestFormatEvent_UnknownEvent(t *testing.T) {
	now := time.Now()
	event := &history.HistoryEvent{
		EventId:   5,
		EventTime: &now,
		EventType: enums.EVENT_TYPE_UNSPECIFIED, // Unknown event type
		Attributes: &history.HistoryEvent_ActivityTaskCanceledEventAttributes{
			ActivityTaskCanceledEventAttributes: &history.ActivityTaskCanceledEventAttributes{},
		},
	}

	formattedEvent, keep := FormatEvent(event, false)
	assert.False(t, keep)
	assert.Equal(t, Event{}, formattedEvent)
}

