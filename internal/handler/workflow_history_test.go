package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
)

// MockTemporalClient is a mock of client.Client
type MockTemporalClient struct {
	client.Client
	mock.Mock
}

func (m *MockTemporalClient) GetWorkflowHistory(ctx context.Context, workflowID string, runID string, isLongPoll bool, filterType enums.HistoryEventFilterType) client.HistoryEventIterator {
	args := m.Called(ctx, workflowID, runID, isLongPoll, filterType)
	return args.Get(0).(client.HistoryEventIterator)
}

// MockHistoryEventIterator is a mock of client.HistoryEventIterator
type MockHistoryEventIterator struct {
	client.HistoryEventIterator
	mock.Mock
}

func (m *MockHistoryEventIterator) HasNext() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockHistoryEventIterator) Next() (*history.HistoryEvent, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*history.HistoryEvent), args.Error(1)
}

func TestGetWorkflowHistoryHandler(t *testing.T) {
	t.Run("Successful history retrieval", func(t *testing.T) {
		mockClient := new(MockTemporalClient)
		mockIterator := new(MockHistoryEventIterator)

		workflowID := "test-workflow-id"
		runID := "test-run-id"
		originalRunID := "original-run-id"

		now := time.Now()
		event1 := &history.HistoryEvent{
			EventId:   1,
			EventTime: &now,
			EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
			Attributes: &history.HistoryEvent_WorkflowExecutionStartedEventAttributes{
				WorkflowExecutionStartedEventAttributes: &history.WorkflowExecutionStartedEventAttributes{
					OriginalExecutionRunId: originalRunID,
				},
			},
		}
		event2 := &history.HistoryEvent{
			EventId:   2,
			EventTime: &now,
			EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED,
		}

		mockIterator.On("HasNext").Return(true).Once()
		mockIterator.On("Next").Return(event1, nil).Once()
		mockIterator.On("HasNext").Return(true).Once()
		mockIterator.On("Next").Return(event2, nil).Once()
		mockIterator.On("HasNext").Return(false).Once()

		mockClient.On("GetWorkflowHistory", mock.Anything, workflowID, runID, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT).Return(mockIterator)

		args := WorkflowHistoryArgs{
			WorkflowID: workflowID,
			RunID:      runID,
		}

		result, err := GetWorkflowHistoryHandler(context.Background(), mockClient, args)

		assert.NoError(t, err)
		assert.Equal(t, workflowID, result.WorkflowID)
		assert.Equal(t, originalRunID, result.RunID)
		assert.Len(t, result.Events, 2)
		assert.Equal(t, int64(1), result.Events[0].EventID)
		assert.Equal(t, "WorkflowExecutionStarted", result.Events[0].Type)
		assert.Equal(t, int64(2), result.Events[1].EventID)
		assert.Equal(t, "WorkflowExecutionCompleted", result.Events[1].Type)
		assert.Equal(t, "Workflow has 2 events. We are examining 2 events.", result.Summary)

		mockClient.AssertExpectations(t)
		mockIterator.AssertExpectations(t)
	})

	t.Run("Error from iterator", func(t *testing.T) {
		mockClient := new(MockTemporalClient)
		mockIterator := new(MockHistoryEventIterator)

		workflowID := "test-workflow-id"
		runID := "test-run-id"
		iteratorError := errors.New("iterator error")

		mockIterator.On("HasNext").Return(true).Once()
		mockIterator.On("Next").Return(nil, iteratorError).Once()

		mockClient.On("GetWorkflowHistory", mock.Anything, workflowID, runID, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT).Return(mockIterator)

		args := WorkflowHistoryArgs{
			WorkflowID: workflowID,
			RunID:      runID,
		}

		_, err := GetWorkflowHistoryHandler(context.Background(), mockClient, args)

		assert.Error(t, err)
		assert.EqualError(t, err, "failed reading history: iterator error")

		mockClient.AssertExpectations(t)
		mockIterator.AssertExpectations(t)
	})

	t.Run("No history events", func(t *testing.T) {
		mockClient := new(MockTemporalClient)
		mockIterator := new(MockHistoryEventIterator)

		workflowID := "test-workflow-id"
		runID := "test-run-id"

		mockIterator.On("HasNext").Return(false).Once()

		mockClient.On("GetWorkflowHistory", mock.Anything, workflowID, runID, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT).Return(mockIterator)

		args := WorkflowHistoryArgs{
			WorkflowID: workflowID,
			RunID:      runID,
		}

		result, err := GetWorkflowHistoryHandler(context.Background(), mockClient, args)

		assert.NoError(t, err)
		assert.Equal(t, workflowID, result.WorkflowID)
		assert.Equal(t, runID, result.RunID) // Should be the one from args as no events were found
		assert.Len(t, result.Events, 0)
		assert.Equal(t, "Workflow has 0 events. We are examining 0 events.", result.Summary)

		mockClient.AssertExpectations(t)
		mockIterator.AssertExpectations(t)
	})

	t.Run("RunID not provided", func(t *testing.T) {
		mockClient := new(MockTemporalClient)
		mockIterator := new(MockHistoryEventIterator)

		workflowID := "test-workflow-id"
		originalRunID := "original-run-id"

		now := time.Now()
		event1 := &history.HistoryEvent{
			EventId:   1,
			EventTime: &now,
			EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
			Attributes: &history.HistoryEvent_WorkflowExecutionStartedEventAttributes{
				WorkflowExecutionStartedEventAttributes: &history.WorkflowExecutionStartedEventAttributes{
					OriginalExecutionRunId: originalRunID,
				},
			},
		}

		mockIterator.On("HasNext").Return(true).Once()
		mockIterator.On("Next").Return(event1, nil).Once()
		mockIterator.On("HasNext").Return(false).Once()

		mockClient.On("GetWorkflowHistory", mock.Anything, workflowID, "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT).Return(mockIterator)

		args := WorkflowHistoryArgs{
			WorkflowID: workflowID,
		}

		result, err := GetWorkflowHistoryHandler(context.Background(), mockClient, args)

		assert.NoError(t, err)
		assert.Equal(t, workflowID, result.WorkflowID)
		assert.Equal(t, originalRunID, result.RunID)
		assert.Len(t, result.Events, 1)
		assert.Equal(t, "Workflow has 1 events. We are examining 1 events.", result.Summary)

		mockClient.AssertExpectations(t)
		mockIterator.AssertExpectations(t)
	})
}
