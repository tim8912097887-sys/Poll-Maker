package vote_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/features/vote"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared"
	httpresponse "github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/response/http_response"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/validation"
	pollv1 "github.com/tim8912097887-sys/Poll-Maker/services/vote_service/proto"
)

// ---------------------------------------------------------------------------
// Router / request helpers
// ---------------------------------------------------------------------------

func setupRouter(t *testing.T, h *vote.Handler) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{
		StructValidator: &validation.StructValidator{
			Validating: validator.New(),
		},
	})
	voteGroup := app.Group("/api/v1/votes")
	h.RegisterRoutes(voteGroup)
	return app
}

func decodeResponse[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	var payload T
	err := json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}


func wireupHandler(t *testing.T, voteRepository *MockVoteRepository, voteCache *MockVoteCache, grpcClient *MockVoteGrpcClient) *vote.Handler {
	t.Helper()
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, handlerOpts))
	slog.SetDefault(logger)

	if grpcClient == nil {
		grpcClient = InitMockVoteGrpcClient()
	}

	voteService := vote.NewService(&vote.ServiceConfig{
		VoteRepository: voteRepository,
		VoteCache:      voteCache,
		GrpcClient:     grpcClient,
		Logger:         logger,
	})

	voteHandler := vote.NewHandler(&vote.HandlerConfig{
		VoteService:  voteService,
		Logger:       logger,
	})
	return voteHandler
}

func postVoteRequest(t *testing.T, app *fiber.App, route string, body any) *http.Response {
	t.Helper()

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, route, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// ---------------------------------------------------------------------------
// Assertion helpers
// ---------------------------------------------------------------------------

func assertValidationError(t *testing.T, resp *http.Response, expectedField, expectedRule string) {
	t.Helper()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	raw := decodeResponse[map[string]any](t, resp)
	errObj, ok := raw["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error field to be an object, got %T", raw["error"])
	}

	if errObj["code"] != "validation_error" {
		t.Fatalf("expected error code %q, got %v", "validation_error", errObj["code"])
	}

	detail, ok := errObj["detail"].([]any)
	if !ok {
		t.Fatalf("expected detail field to be a slice, got %T", errObj["detail"])
	}

	found := false
	for _, d := range detail {
		entry, ok := d.(map[string]any)
		if !ok {
			continue
		}
		if entry["field"] == expectedField && entry["rule"] == expectedRule {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected detail to contain field %q with rule %q, got %v", expectedField, expectedRule, detail)
	}
}

func assertErrorResponse(t *testing.T, resp *http.Response, expectedStatus int, expectedCode string) httpresponse.ErrorResponse {
	t.Helper()

	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}

	errorResponse := decodeResponse[httpresponse.ErrorResponse](t, resp)
	if errorResponse.Error.Code != expectedCode {
		t.Fatalf("expected error code %q, got %q", expectedCode, errorResponse.Error.Code)
	}
	return errorResponse
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCreateVoteValidation(t *testing.T) {
	voteRepository := InitMockVoteRepository()
	voteCache := InitMockVoteCache()
	voteGrpcClient := InitMockVoteGrpcClient()
	voteHandler := wireupHandler(t, voteRepository, voteCache, voteGrpcClient)

	validPollId := uuid.New().String()
	validOptionId := uuid.New().String()

	tests := []struct {
		name          string
		body          map[string]any
		expectedField string
		expectedRule  string
	}{
		{
			name: "sessionId is required",
			body: map[string]any{
				"pollId":   validPollId,
				"optionId": validOptionId,
			},
			expectedField: "SessionId",
			expectedRule:  "required",
		},
		{
			name: "pollId is required",
			body: map[string]any{
				"sessionId": "session-abc",
				"optionId":  validOptionId,
			},
			expectedField: "PollId",
			expectedRule:  "required",
		},
		{
			name: "pollId must be a uuid",
			body: map[string]any{
				"sessionId": "session-abc",
				"pollId":    "not-a-uuid",
				"optionId":  validOptionId,
			},
			expectedField: "PollId",
			expectedRule:  "uuid",
		},
		{
			name: "optionId is required",
			body: map[string]any{
				"sessionId": "session-abc",
				"pollId":    validPollId,
			},
			expectedField: "OptionId",
			expectedRule:  "required",
		},
		{
			name: "optionId must be a uuid",
			body: map[string]any{
				"sessionId": "session-abc",
				"pollId":    validPollId,
				"optionId":  "not-a-uuid",
			},
			expectedField: "OptionId",
			expectedRule:  "uuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupRouter(t, voteHandler)
			resp := postVoteRequest(t, app, "/api/v1/votes", tt.body)
			assertValidationError(t, resp, tt.expectedField, tt.expectedRule)
		})
	}
}

func TestCreateVoteSuccess(t *testing.T) {
	voteRepository := InitMockVoteRepository()
	voteCache := InitMockVoteCache()
	voteGrpcClient := InitMockVoteGrpcClient()
	voteHandler := wireupHandler(t, voteRepository, voteCache, voteGrpcClient)

	pollId := uuid.New().String()
	optionId := uuid.New().String()
	sessionId := "session-abc"

	app := setupRouter(t, voteHandler)
	resp := postVoteRequest(t, app, "/api/v1/votes", map[string]any{
		"sessionId": sessionId,
		"pollId":    pollId,
		"optionId":  optionId,
	})

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	successResponse := decodeResponse[httpresponse.SuccessResponse](t, resp)
	if successResponse.State != "success" {
		t.Fatalf("expected state %q, got %q", "success", successResponse.State)
	}
	if successResponse.Error != nil {
		t.Fatalf("expected no error, got %v", successResponse.Error)
	}

	data, ok := successResponse.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data to be an object, got %T", successResponse.Data)
	}

	voteData, ok := data["vote"].(map[string]any)
	if !ok {
		t.Fatalf("expected vote field to be an object, got %T", data["vote"])
	}

	if voteData["sessionId"] != sessionId {
		t.Fatalf("expected sessionId %q, got %q", sessionId, voteData["sessionId"])
	}
	if voteData["pollId"] != pollId {
		t.Fatalf("expected pollId %q, got %q", pollId, voteData["pollId"])
	}
	if voteData["optionId"] != optionId {
		t.Fatalf("expected optionId %q, got %q", optionId, voteData["optionId"])
	}
	if voteData["id"] == "" || voteData["id"] == nil {
		t.Fatalf("expected a non-empty id, got %v", voteData["id"])
	}
}

func TestCreateVoteAlreadyVoted(t *testing.T) {
	voteRepository := InitMockVoteRepository()
	voteCache := InitMockVoteCache()
	voteCache.HasVotedFunc = func(ctx context.Context, pollID, sessionID string) (bool, error) {
		return true, nil
	}
	voteGrpcClient := InitMockVoteGrpcClient()
	voteHandler := wireupHandler(t, voteRepository, voteCache, voteGrpcClient)

	app := setupRouter(t, voteHandler)
	resp := postVoteRequest(t, app, "/api/v1/votes", map[string]any{
		"sessionId": "session-abc",
		"pollId":    uuid.New().String(),
		"optionId":  uuid.New().String(),
	})

	errorResponse := assertErrorResponse(t, resp, http.StatusBadRequest, "AREADY_VOTED")
	if errorResponse.Error.Message != shared.ErrAlreadyVoted.Error() {
		t.Fatalf("expected error message %q, got %q", shared.ErrAlreadyVoted.Error(), errorResponse.Error.Message)
	}
}

func TestCreateVoteInvalidOption(t *testing.T) {
	voteRepository := InitMockVoteRepository()
	voteCache := InitMockVoteCache()
	voteCache.IsValidOptionFunc = func(ctx context.Context, pollID, optionID string) (bool, error) {
		return false, nil
	}
	voteGrpcClient := InitMockVoteGrpcClient()
	voteHandler := wireupHandler(t, voteRepository, voteCache, voteGrpcClient)

	app := setupRouter(t, voteHandler)
	resp := postVoteRequest(t, app, "/api/v1/votes", map[string]any{
		"sessionId": "session-abc",
		"pollId":    uuid.New().String(),
		"optionId":  uuid.New().String(),
	})

	errorResponse := assertErrorResponse(t, resp, http.StatusBadRequest, "INVALID_OPTION")
	if errorResponse.Error.Message != shared.ErrInvalidOption.Error() {
		t.Fatalf("expected error message %q, got %q", shared.ErrInvalidOption.Error(), errorResponse.Error.Message)
	}
}

func TestCreateVoteInternalServerError(t *testing.T) {
	voteRepository := InitMockVoteRepository()
	voteRepository.CreateVoteFunc = func(ctx context.Context, id string, vote types.CreateVoteSchema, createVoteEvent types.CreateVoteEvent, expiredAt time.Time) (types.CreateVoteResponse, error) {
		return types.CreateVoteResponse{}, errors.New("database unavailable")
	}
	voteCache := InitMockVoteCache()
	voteGrpcClient := InitMockVoteGrpcClient()
	voteHandler := wireupHandler(t, voteRepository, voteCache, voteGrpcClient)

	app := setupRouter(t, voteHandler)
	resp := postVoteRequest(t, app, "/api/v1/votes", map[string]any{
		"sessionId": "session-abc",
		"pollId":    uuid.New().String(),
		"optionId":  uuid.New().String(),
	})

	errorResponse := assertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
	if errorResponse.Error.Message != "internal server error" {
		t.Fatalf("expected error message %q, got %q", "internal server error", errorResponse.Error.Message)
	}
}

func TestCreateVoteGrpcClientResponses(t *testing.T) {
	pollID := uuid.New().String()
	optionID := uuid.New().String()
	sessionID := "session-abc"

	tests := []struct {
		name           string
		grpcResponse   *pollv1.ValidatePollResponse
		grpcErr        error
		expectedStatus int
		expectedCode   string
		expectedMsg    string
		assertResponse func(*testing.T, *http.Response)
	}{
		{
			name:           "grpc client returns valid response",
			grpcResponse:   &pollv1.ValidatePollResponse{IsValid: true},
			expectedStatus: http.StatusOK,
			assertResponse: func(t *testing.T, resp *http.Response) {
				t.Helper()
				successResponse := decodeResponse[httpresponse.SuccessResponse](t, resp)
				if successResponse.State != "success" {
					t.Fatalf("expected state %q, got %q", "success", successResponse.State)
				}
			},
		},
		{
			name:           "grpc client returns poll not found",
			grpcResponse:   &pollv1.ValidatePollResponse{IsValid: false, Reason: pollv1.ValidatePollResponse_POLL_NOT_FOUND},
			expectedStatus: http.StatusNotFound,
			expectedCode:   "POLL_NOT_FOUND",
			expectedMsg:    shared.ErrPollNotFound.Error(),
		},
		{
			name:           "grpc client returns poll expired",
			grpcResponse:   &pollv1.ValidatePollResponse{IsValid: false, Reason: pollv1.ValidatePollResponse_POLL_EXPIRED},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "POLL_EXPIRED",
			expectedMsg:    shared.ErrPollExpired.Error(),
		},
		{
			name:           "grpc client returns poll not started",
			grpcResponse:   &pollv1.ValidatePollResponse{IsValid: false, Reason: pollv1.ValidatePollResponse_POLL_NOT_STARTED},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "POLL_NOT_STARTED",
			expectedMsg:    shared.ErrPollNotStarted.Error(),
		},
		{
			name:           "grpc client returns poll closed",
			grpcResponse:   &pollv1.ValidatePollResponse{IsValid: false, Reason: pollv1.ValidatePollResponse_POLL_CLOSED},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "POLL_CLOSED",
			expectedMsg:    shared.ErrPollClosed.Error(),
		},
		{
			name:           "grpc client returns unspecified reason",
			grpcResponse:   &pollv1.ValidatePollResponse{IsValid: false, Reason: pollv1.ValidatePollResponse_REASON_UNSPECIFIED},
			expectedStatus: http.StatusNotFound,
			expectedCode:   "POLL_NOT_FOUND",
			expectedMsg:    shared.ErrPollNotFound.Error(),
		},
		{
			name:           "grpc client returns an error",
			grpcErr:        errors.New("grpc unavailable"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
			expectedMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			voteRepository := InitMockVoteRepository()
			voteCache := InitMockVoteCache()
			voteCache.GetPollMetaFunc = func(ctx context.Context, pollID string) (*types.PollMeta, error) {
				return nil, shared.ErrPollNotFound
			}

			grpcClient := &MockVoteGrpcClient{
				ValidatePollForVotingFunc: func(ctx context.Context, pollID string) (*pollv1.ValidatePollResponse, error) {
					return tt.grpcResponse, tt.grpcErr
				},
			}

			voteHandler := wireupHandler(t, voteRepository, voteCache, grpcClient)
			app := setupRouter(t, voteHandler)
			resp := postVoteRequest(t, app, "/api/v1/votes", map[string]any{
				"sessionId": sessionID,
				"pollId":    pollID,
				"optionId":  optionID,
			})

			if tt.assertResponse != nil {
				tt.assertResponse(t, resp)
				return
			}

			errorResponse := assertErrorResponse(t, resp, tt.expectedStatus, tt.expectedCode)
			if errorResponse.Error.Message != tt.expectedMsg {
				t.Fatalf("expected error message %q, got %q", tt.expectedMsg, errorResponse.Error.Message)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type MockVoteRepository struct {
	CreateVoteFunc func(ctx context.Context,id string, vote types.CreateVoteSchema, createVoteEvent types.CreateVoteEvent,expiredAt time.Time) (types.CreateVoteResponse, error)
    GetOutboxEventsFunc func(ctx context.Context,limit int) ([]types.CreateVoteEvent, error)
    UpdateOutboxEventsFunc func(ctx context.Context, eventIds []string) error
}

func InitMockVoteRepository() *MockVoteRepository {
	return &MockVoteRepository{
		CreateVoteFunc: func(ctx context.Context,id string, vote types.CreateVoteSchema, createVoteEvent types.CreateVoteEvent,expiredAt time.Time) (types.CreateVoteResponse, error) {
			return types.CreateVoteResponse{
				Id:        uuid.MustParse(id),
				SessionId: vote.SessionId,
				PollId:    uuid.MustParse(vote.PollId),
				OptionId:  uuid.MustParse(vote.OptionId),
			}, nil
		},
		GetOutboxEventsFunc: func(ctx context.Context,limit int) ([]types.CreateVoteEvent, error) {
			return []types.CreateVoteEvent{
				{
					EventId:   uuid.NewString(),
					PollId:    uuid.NewString(),
					OptionId:  uuid.NewString(),
					VotedAt:   time.Now().Format(time.RFC3339),
				},
			}, nil
		},
		UpdateOutboxEventsFunc: func(ctx context.Context, eventIds []string) error {
			return nil
		},
	}
}

func (m *MockVoteRepository)CreateVote(ctx context.Context,id string, vote types.CreateVoteSchema, createVoteEvent types.CreateVoteEvent,expiredAt time.Time) (types.CreateVoteResponse, error) {
	return m.CreateVoteFunc(ctx, id, vote, createVoteEvent, expiredAt)
}

func (m *MockVoteRepository) GetOutboxEvents(ctx context.Context,limit int) ([]types.CreateVoteEvent, error) {
	return m.GetOutboxEvents(ctx,limit)
}

func (m *MockVoteRepository) UpdateOutboxEvents(ctx context.Context, eventIds []string) error {
	return m.UpdateOutboxEventsFunc(ctx, eventIds)
}

type MockVoteGrpcClient struct {
	ValidatePollForVotingFunc func(ctx context.Context, pollID string) (*pollv1.ValidatePollResponse, error)
}

func InitMockVoteGrpcClient() *MockVoteGrpcClient {
	return &MockVoteGrpcClient{
		ValidatePollForVotingFunc: func(ctx context.Context, pollID string) (*pollv1.ValidatePollResponse, error) {
			return &pollv1.ValidatePollResponse{}, nil
		},
	}
}

func (m *MockVoteGrpcClient) ValidatePollForVoting(ctx context.Context, pollID string) (*pollv1.ValidatePollResponse, error) {
	return m.ValidatePollForVotingFunc(ctx, pollID)
}

type MockVoteCache struct {
	HasVotedFunc        func(ctx context.Context, pollID, sessionID string) (bool, error)
	MarkVotedFunc       func(ctx context.Context, pollID, sessionID string, expiredAt time.Time) error
	GetPollMetaFunc     func(ctx context.Context, pollID string) (*types.PollMeta, error)
	IsValidOptionFunc   func(ctx context.Context, pollID, optionID string) (bool, error)
	DeleteVoteCacheFunc func(ctx context.Context, pollID string) error
}

func InitMockVoteCache() *MockVoteCache {
	return &MockVoteCache{
		HasVotedFunc: func(ctx context.Context, pollID, sessionID string) (bool, error) {
			return false, nil
		},
		MarkVotedFunc: func(ctx context.Context, pollID, sessionID string, expiredAt time.Time) error {
			return nil
		},
		GetPollMetaFunc: func(ctx context.Context, pollID string) (*types.PollMeta, error) {
			return &types.PollMeta{
				StartedAt: time.Now(),
				ExpiredAt: time.Now().Add(time.Hour),
				IsPrivate: false,
			}, nil
		},
		IsValidOptionFunc: func(ctx context.Context, pollID, optionID string) (bool, error) {
			return true, nil
		},
		DeleteVoteCacheFunc: func(ctx context.Context, pollID string) error {
			return nil
		},
	}
}

func (m *MockVoteCache) HasVoted(ctx context.Context, pollID, sessionID string) (bool, error) {
	return m.HasVotedFunc(ctx, pollID, sessionID)
}

func (m *MockVoteCache) MarkVoted(ctx context.Context, pollID, sessionID string, expiredAt time.Time) error {
	return m.MarkVotedFunc(ctx, pollID, sessionID, expiredAt)
}

func (m *MockVoteCache) GetPollMeta(ctx context.Context, pollID string) (*types.PollMeta, error) {
	return m.GetPollMetaFunc(ctx, pollID)
}

func (m *MockVoteCache) IsValidOption(ctx context.Context, pollID, optionID string) (bool, error) {
	return m.IsValidOptionFunc(ctx, pollID, optionID)
}

func (m *MockVoteCache) DeleteVoteCache(ctx context.Context, pollID string) error {
	return m.DeleteVoteCacheFunc(ctx, pollID)
}
