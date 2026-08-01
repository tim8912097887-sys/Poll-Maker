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
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/middlewares"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/response"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/types"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/validation"
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

func wireupHandler(t *testing.T, voteRepository *MockVoteRepository, voteCache *MockVoteCache) *vote.Handler {
	t.Helper()
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, handlerOpts))
	slog.SetDefault(logger)

	voteService := vote.NewService(&vote.ServiceConfig{
		VoteRepository: voteRepository,
		VoteCache:      voteCache,
	})

	voteHandler := vote.NewHandler(&vote.HandlerConfig{
		VoteService:  voteService,
		Logger:       logger,
		ErrorHandler: middlewares.ErrorHandlerMiddleware(),
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

func assertErrorResponse(t *testing.T, resp *http.Response, expectedStatus int, expectedCode string) response.ErrorResponse {
	t.Helper()

	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}

	errorResponse := decodeResponse[response.ErrorResponse](t, resp)
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
	voteHandler := wireupHandler(t, voteRepository, voteCache)

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
	voteHandler := wireupHandler(t, voteRepository, voteCache)

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

	successResponse := decodeResponse[response.SuccessResponse](t, resp)
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
	voteHandler := wireupHandler(t, voteRepository, voteCache)

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
	voteHandler := wireupHandler(t, voteRepository, voteCache)

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
	voteRepository.CreateVoteFunc = func(ctx context.Context, id string, v types.CreateVoteSchema) (types.CreateVoteResponse, error) {
		return types.CreateVoteResponse{}, errors.New("database unavailable")
	}
	voteCache := InitMockVoteCache()
	voteHandler := wireupHandler(t, voteRepository, voteCache)

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

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type MockVoteRepository struct {
	CreateVoteFunc func(ctx context.Context, id string, vote types.CreateVoteSchema) (types.CreateVoteResponse, error)
}

func InitMockVoteRepository() *MockVoteRepository {
	return &MockVoteRepository{
		CreateVoteFunc: func(ctx context.Context, id string, vote types.CreateVoteSchema) (types.CreateVoteResponse, error) {
			return types.CreateVoteResponse{
				Id:        uuid.MustParse(id),
				SessionId: vote.SessionId,
				PollId:    uuid.MustParse(vote.PollId),
				OptionId:  uuid.MustParse(vote.OptionId),
			}, nil
		},
	}
}

func (m *MockVoteRepository) CreateVote(ctx context.Context, id string, vote types.CreateVoteSchema) (types.CreateVoteResponse, error) {
	return m.CreateVoteFunc(ctx, id, vote)
}

type MockVoteCache struct {
	HasVotedFunc      func(ctx context.Context, pollID, sessionID string) (bool, error)
	MarkVotedFunc     func(ctx context.Context, pollID, sessionID string, expiredAt time.Time) error
	GetPollMetaFunc   func(ctx context.Context, pollID string) (*types.PollMeta, error)
	IsValidOptionFunc func(ctx context.Context, pollID, optionID string) (bool, error)
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
