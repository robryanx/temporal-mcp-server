package handler_test

import (
	"context"
	"testing"
	"time"

	"github.com/robryanx/mcp-temporal-server/internal/handler"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/failure/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/mocks"
	"go.temporal.io/sdk/testsuite"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockHistoryIterator struct {
	mock.Mock
}

func (m *MockHistoryIterator) HasNext() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockHistoryIterator) Next() (*history.HistoryEvent, error) {
	args := m.Called()
	return args.Get(0).(*history.HistoryEvent), args.Error(1)
}

type HandlerSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	client *mocks.Client
}

func (s *HandlerSuite) SetupTest() {
	s.client = &mocks.Client{}
}

func (s *HandlerSuite) TestGetFailedWorkflowsHandler() {
	now := time.Now()
	historyEvent := &history.HistoryEvent{
		EventId:   1,
		EventTime: &now,
		EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED,
		Attributes: &history.HistoryEvent_WorkflowExecutionFailedEventAttributes{
			WorkflowExecutionFailedEventAttributes: &history.WorkflowExecutionFailedEventAttributes{
				Failure: &failure.Failure{
					Message: "test error",
				},
			},
		},
	}

	historyIterator := &MockHistoryIterator{}
	historyIterator.On("HasNext").Return(true).Once()
	historyIterator.On("HasNext").Return(false)
	historyIterator.On("Next").Return(historyEvent, nil).Once()

	s.client.On("GetWorkflowHistory", mock.Anything, "test-workflow-id", "test-run-id", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT).Return(historyIterator)

	s.client.On("ListOpenWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.ListOpenWorkflowExecutionsResponse{
		Executions: []*workflow.WorkflowExecutionInfo{
			{
				Execution: &common.WorkflowExecution{
					WorkflowId: "test-workflow-id",
					RunId:      "test-run-id",
				},
			},
		},
	}, nil)

	resp, err := handler.GetFailedWorkflowsHandler(context.Background(), s.client)
	s.NoError(err)
	s.Len(resp.Workflows, 1)
	s.Equal("test-workflow-id", resp.Workflows[0].WorkflowID)
	s.Equal("test-run-id", resp.Workflows[0].RunID)
	s.Equal("test error", resp.Workflows[0].Error)
	s.Len(resp.Workflows[0].Summary, 1)
}

func (s *HandlerSuite) TestFormatEvent_WorkflowExecutionFailed() {
	now := time.Now()
	historyEvent := &history.HistoryEvent{
		EventId:   1,
		EventTime: &now,
		EventType: enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED,
		Attributes: &history.HistoryEvent_WorkflowExecutionFailedEventAttributes{
			WorkflowExecutionFailedEventAttributes: &history.WorkflowExecutionFailedEventAttributes{
				Failure: &failure.Failure{
					Message: "test error message",
				},
			},
		},
	}

	formattedEvent, keep := handler.FormatEvent(historyEvent, true)
	s.True(keep)
	s.Equal("test error message", formattedEvent.Details["error"])
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}