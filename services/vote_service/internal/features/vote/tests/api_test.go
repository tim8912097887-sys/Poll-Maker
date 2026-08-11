package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/features/vote"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/features/vote/tests/helper"
	"github.com/tim8912097887-sys/Poll-Maker/services/vote_service/internal/shared/validation"
	pollv1 "github.com/tim8912097887-sys/Poll-Maker/services/vote_service/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	voteMetaKeyFmt    = "poll:%s:meta"
	voteOptionsKeyFmt = "poll:%s:options"
	voteSetKeyFmt     = "poll:%s:voted"
)

type mockGrpcClient struct {
	response *pollv1.ValidatePollResponse
	err      error
	called   int
}

func (m *mockGrpcClient) ValidatePollForVoting(ctx context.Context, pollID string) (*pollv1.ValidatePollResponse, error) {
	m.called++
	return m.response, m.err
}

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

func postVoteRequest(t *testing.T, app *fiber.App, body any) *http.Response {
	t.Helper()
	requestBody, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/votes", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

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
		t.Fatalf("expected validation_error code, got %v", errObj["code"])
	}

	detail, ok := errObj["detail"].([]any)
	if !ok {
		t.Fatalf("expected detail field to be a slice, got %T", errObj["detail"])
	}

	found := false
	for _, entry := range detail {
		field, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if field["field"] == expectedField && field["rule"] == expectedRule {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected validation error for %s=%s, got %v", expectedField, expectedRule, detail)
	}
}

func assertErrorResponse(t *testing.T, resp *http.Response, expectedStatus int, expectedCode string) map[string]any {
	t.Helper()
	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}

	raw := decodeResponse[map[string]any](t, resp)
	errObj, ok := raw["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error field to be an object, got %T", raw["error"])
	}
	if errObj["code"] != expectedCode {
		t.Fatalf("expected error code %q, got %q", expectedCode, errObj["code"])
	}
	return errObj
}

func newIntegrationHandler(t *testing.T, pool *pgxpool.Pool, rdb *redis.Client, grpcClient vote.GrpcClient) *vote.Handler {
	t.Helper()

	handlerOpts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, handlerOpts))
	slog.SetDefault(logger)

	voteService := vote.NewService(&vote.ServiceConfig{
		VoteRepository: vote.NewRepository(pool),
		VoteCache:      vote.NewCache(vote.CacheConfig{CacheClient: rdb}),
		GrpcClient:     grpcClient,
		Logger:         logger,
	})
	return vote.NewHandler(&vote.HandlerConfig{
		VoteService: voteService,
		Logger:      logger,
	})
}

func seedPollAndOption(t *testing.T, ctx context.Context, pool *pgxpool.Pool, pollID, optionID string, startedAt, expiredAt time.Time) {
	t.Helper()

	_, err := pool.Exec(ctx, `INSERT INTO polls (id, title, is_private, creator_session, started_at, expired_at) VALUES ($1, $2, $3, $4, $5, $6)`, pollID, "integration poll", false, "creator-session", startedAt, expiredAt)
	if err != nil {
		t.Fatal(err)
	}

	_, err = pool.Exec(ctx, `INSERT INTO poll_options (id, poll_id, option_text) VALUES ($1, $2, $3)`, optionID, pollID, "option text")
	if err != nil {
		t.Fatal(err)
	}
}

func seedPollCache(t *testing.T, ctx context.Context, rdb *redis.Client, pollID, optionID string, startedAt, expiredAt time.Time) {
	t.Helper()

	metaKey := fmt.Sprintf(voteMetaKeyFmt, pollID)
	optionKey := fmt.Sprintf(voteOptionsKeyFmt, pollID)
	pipe := rdb.TxPipeline()
	pipe.HSet(ctx, metaKey, "StartedAt", startedAt.Format(time.RFC3339), "ExpiredAt", expiredAt.Format(time.RFC3339), "IsPrivate", "false")
	pipe.SAdd(ctx, optionKey, optionID)
	_, err := pipe.Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func countVotes(t *testing.T, ctx context.Context, pool *pgxpool.Pool, pollID, sessionID, optionID string) int {
	t.Helper()
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM votes WHERE poll_id = $1 AND session_id = $2 AND option_id = $3`, pollID, sessionID, optionID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func countOutboxEvents(t *testing.T, ctx context.Context, pool *pgxpool.Pool, pollID, optionID string) int {
	t.Helper()
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM outbox_events WHERE poll_id = $1 AND option_id = $2 AND status = 'pending'`, pollID, optionID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func cacheHasVoted(t *testing.T, ctx context.Context, rdb *redis.Client, pollID, sessionID string) bool {
	t.Helper()
	key := fmt.Sprintf(voteSetKeyFmt, pollID)
	member, err := rdb.SIsMember(ctx, key, sessionID).Result()
	if err != nil {
		t.Fatal(err)
	}
	return member
}

func cacheVoteTTL(t *testing.T, ctx context.Context, rdb *redis.Client, pollID string) time.Duration {
	t.Helper()
	key := fmt.Sprintf(voteSetKeyFmt, pollID)
	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	return ttl
}

func TestCreateVoteValidationIntegration(t *testing.T) {
	pool, rdb, _, cleanup := helper.InitIntegrationDeps(t)
	defer cleanup()

	mockGrpc := &mockGrpcClient{response: &pollv1.ValidatePollResponse{IsValid: true, ExpiredAt: timestamppb.New(time.Now().Add(time.Hour))}}
	handler := newIntegrationHandler(t, pool, rdb, mockGrpc)
	app := setupRouter(t, handler)

	validPollID := uuid.NewString()
	validOptionID := uuid.NewString()

	tests := []struct {
		name          string
		body          map[string]any
		expectedField string
		expectedRule  string
	}{
		{name: "missing sessionId", body: map[string]any{"pollId": validPollID, "optionId": validOptionID}, expectedField: "SessionId", expectedRule: "required"},
		{name: "missing pollId", body: map[string]any{"sessionId": "session-1", "optionId": validOptionID}, expectedField: "PollId", expectedRule: "required"},
		{name: "invalid pollId", body: map[string]any{"sessionId": "session-1", "pollId": "not-uuid", "optionId": validOptionID}, expectedField: "PollId", expectedRule: "uuid"},
		{name: "missing optionId", body: map[string]any{"sessionId": "session-1", "pollId": validPollID}, expectedField: "OptionId", expectedRule: "required"},
		{name: "invalid optionId", body: map[string]any{"sessionId": "session-1", "pollId": validPollID, "optionId": "not-uuid"}, expectedField: "OptionId", expectedRule: "uuid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := postVoteRequest(t, app, tt.body)
			assertValidationError(t, resp, tt.expectedField, tt.expectedRule)
		})
	}
}

func TestCreateVoteSuccessIntegration(t *testing.T) {
	pool, rdb, ctx, cleanup := helper.InitIntegrationDeps(t)
	defer cleanup()

	pollID := uuid.NewString()
	optionID := uuid.NewString()
	sessionID := "session-success"
	startedAt := time.Now().Add(-time.Hour)
	expiredAt := time.Now().Add(time.Hour)

	defer helper.CleanupPollData(t, ctx, pool, rdb, pollID)
	seedPollAndOption(t, ctx, pool, pollID, optionID, startedAt, expiredAt)
	seedPollCache(t, ctx, rdb, pollID, optionID, startedAt, expiredAt)

	mockGrpc := &mockGrpcClient{response: &pollv1.ValidatePollResponse{IsValid: true, ExpiredAt: timestamppb.New(expiredAt)}}
	handler := newIntegrationHandler(t, pool, rdb, mockGrpc)
	app := setupRouter(t, handler)

	resp := postVoteRequest(t, app, map[string]any{"sessionId": sessionID, "pollId": pollID, "optionId": optionID})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, resp.StatusCode)
	}

	success := decodeResponse[map[string]any](t, resp)
	if success["state"] != "success" {
		t.Fatalf("expected success state got %v", success["state"])
	}

	if countVotes(t, ctx, pool, pollID, sessionID, optionID) != 1 {
		t.Fatal("expected vote row to be created")
	}
	if countOutboxEvents(t, ctx, pool, pollID, optionID) != 1 {
		t.Fatal("expected pending outbox event to be created")
	}
	if !cacheHasVoted(t, ctx, rdb, pollID, sessionID) {
		t.Fatal("expected vote cache to contain session membership")
	}
	if ttl := cacheVoteTTL(t, ctx, rdb, pollID); ttl <= 0 {
		t.Fatalf("expected vote cache TTL to be positive, got %v", ttl)
	}
}

func TestCreateVoteAlreadyVotedIntegration(t *testing.T) {
	pool, rdb, ctx, cleanup := helper.InitIntegrationDeps(t)
	defer cleanup()

	pollID := uuid.NewString()
	optionID := uuid.NewString()
	sessionID := "session-already-voted"
	startedAt := time.Now().Add(-time.Hour)
	expiredAt := time.Now().Add(time.Hour)

	defer helper.CleanupPollData(t, ctx, pool, rdb, pollID)
	seedPollAndOption(t, ctx, pool, pollID, optionID, startedAt, expiredAt)
	seedPollCache(t, ctx, rdb, pollID, optionID, startedAt, expiredAt)
	if err := rdb.SAdd(ctx, fmt.Sprintf(voteSetKeyFmt, pollID), sessionID).Err(); err != nil {
		t.Fatal(err)
	}

	mockGrpc := &mockGrpcClient{response: &pollv1.ValidatePollResponse{IsValid: true, ExpiredAt: timestamppb.New(expiredAt)}}
	handler := newIntegrationHandler(t, pool, rdb, mockGrpc)
	app := setupRouter(t, handler)

	before := countVotes(t, ctx, pool, pollID, sessionID, optionID)
	resp := postVoteRequest(t, app, map[string]any{"sessionId": sessionID, "pollId": pollID, "optionId": optionID})
	assertErrorResponse(t, resp, http.StatusBadRequest, "AREADY_VOTED")
	if countVotes(t, ctx, pool, pollID, sessionID, optionID) != before {
		t.Fatal("expected no vote row to be created after duplicate vote")
	}
}

func TestCreateVoteFallbackToGrpcIntegrationFailure(t *testing.T) {
	pool, rdb, ctx, cleanup := helper.InitIntegrationDeps(t)
	defer cleanup()

	pollID := uuid.NewString()
	optionID := uuid.NewString()
	sessionID := "session-grpc-fallback"
	startedAt := time.Now().Add(-time.Hour)
	expiredAt := time.Now().Add(time.Hour)

	defer helper.CleanupPollData(t, ctx, pool, rdb, pollID)
	seedPollAndOption(t, ctx, pool, pollID, optionID, startedAt, expiredAt)
	
	test := []struct {
		name          string
		grpcResponse  *pollv1.ValidatePollResponse
		expectedErrorCode string
		expectedStatus int
	}{
		{name: "poll not found", grpcResponse: &pollv1.ValidatePollResponse{IsValid: false, Reason: pollv1.ValidatePollResponse_POLL_NOT_FOUND}, expectedErrorCode: "POLL_NOT_FOUND", expectedStatus: http.StatusNotFound},
		{name: "poll expired", grpcResponse: &pollv1.ValidatePollResponse{IsValid: false, Reason: pollv1.ValidatePollResponse_POLL_EXPIRED}, expectedErrorCode: "POLL_EXPIRED", expectedStatus: http.StatusBadRequest},
		{name: "poll not started", grpcResponse: &pollv1.ValidatePollResponse{IsValid: false, Reason: pollv1.ValidatePollResponse_POLL_NOT_STARTED}, expectedErrorCode: "POLL_NOT_STARTED", expectedStatus: http.StatusBadRequest},
		{name: "poll server error", grpcResponse: &pollv1.ValidatePollResponse{IsValid: false, Reason: pollv1.ValidatePollResponse_REASON_UNSPECIFIED}, expectedErrorCode: "POLL_NOT_FOUND", expectedStatus: http.StatusNotFound},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
	
			    mockGrpc := &mockGrpcClient{response: tt.grpcResponse}
				handler := newIntegrationHandler(t, pool, rdb, mockGrpc)
				app := setupRouter(t, handler)

				resp := postVoteRequest(t, app, map[string]any{"sessionId": sessionID, "pollId": pollID, "optionId": optionID})
				if resp.StatusCode != tt.expectedStatus {
					t.Fatalf("expected status %d got %d", tt.expectedStatus, resp.StatusCode)
				}
				if mockGrpc.called == 0 {
					t.Fatal("expected fallback to grpc ValidatePollForVoting")
				}
				if countVotes(t, ctx, pool, pollID, sessionID, optionID) != 0 {
					t.Fatal("expected vote row not to be created after grpc fallback failure")
				}
				if cacheHasVoted(t, ctx, rdb, pollID, sessionID) {
					t.Fatal("expected vote cache not to contain session membership after grpc fallback failure")
				}
		})
	}
}

func TestCreateVoteFallbackToGrpcIntegrationSuccess(t *testing.T) {
	pool, rdb, ctx, cleanup := helper.InitIntegrationDeps(t)
	defer cleanup()

	pollID := uuid.NewString()
	optionID := uuid.NewString()
	sessionID := "session-grpc-fallback"
	startedAt := time.Now().Add(-time.Hour)
	expiredAt := time.Now().Add(time.Hour)

	defer helper.CleanupPollData(t, ctx, pool, rdb, pollID)
	seedPollAndOption(t, ctx, pool, pollID, optionID, startedAt, expiredAt)

	mockGrpc := &mockGrpcClient{response: &pollv1.ValidatePollResponse{IsValid: true, ExpiredAt: timestamppb.New(expiredAt)}}
	handler := newIntegrationHandler(t, pool, rdb, mockGrpc)
	app := setupRouter(t, handler)

	resp := postVoteRequest(t, app, map[string]any{"sessionId": sessionID, "pollId": pollID, "optionId": optionID})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, resp.StatusCode)
	}
	if mockGrpc.called == 0 {
		t.Fatal("expected fallback to grpc ValidatePollForVoting")
	}
	countVotes := countVotes(t, ctx, pool, pollID, sessionID, optionID)
	if countVotes != 1 {
		t.Fatal("expected vote row created after grpc fallback")
	}
	if !cacheHasVoted(t, ctx, rdb, pollID, sessionID) {
		t.Fatal("expected vote cache to contain session membership after grpc fallback")
	}
}

func TestCreateVoteFoundVoteInCacheIntegration(t *testing.T) {

	pool, rdb, ctx, cleanup := helper.InitIntegrationDeps(t)
	defer cleanup()


	mockGrpc := &mockGrpcClient{response: &pollv1.ValidatePollResponse{IsValid: true, ExpiredAt: timestamppb.New(time.Now().Add(time.Hour))}}
	handler := newIntegrationHandler(t, pool, rdb, mockGrpc)
	app := setupRouter(t, handler)

	test := [] struct {
		name string
		startedAt time.Time
		expiredAt time.Time
		expectedErrorCode string
	}{
		{name: "poll not started", startedAt: time.Now().Add(time.Hour), expiredAt: time.Now().Add(2*time.Hour), expectedErrorCode: "POLL_NOT_STARTED"},
		{name: "poll expired", startedAt: time.Now().Add(-2*time.Hour), expiredAt: time.Now().Add(time.Microsecond), expectedErrorCode: "POLL_EXPIRED"},
		{name: "poll option invalid", startedAt: time.Now().Add(-time.Hour), expiredAt: time.Now().Add(time.Hour), expectedErrorCode: "INVALID_OPTION"},
	}
	
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			pollID := uuid.NewString()
			optionID := uuid.NewString()
			sessionID := "session-invalid-option"
			metaKey := fmt.Sprintf(voteMetaKeyFmt, pollID)
			if err := rdb.HSet(ctx, metaKey, "StartedAt", tt.startedAt.Format(time.RFC3339), "ExpiredAt", tt.expiredAt.Format(time.RFC3339), "IsPrivate", "false").Err(); err != nil {
				t.Fatal(err)
			}
		
			resp := postVoteRequest(t, app, map[string]any{"sessionId": sessionID, "pollId": pollID, "optionId": optionID})
			assertErrorResponse(t, resp, http.StatusBadRequest, tt.expectedErrorCode)
			if countVotes(t, ctx, pool, pollID, sessionID, optionID) != 0 {
				t.Fatal("expected no vote row to be created when option is invalid")
			}
		})
	}
}
 