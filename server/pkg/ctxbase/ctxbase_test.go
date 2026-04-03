package ctxbase_test

import (
	"context"
	"testing"

	"github.com/sonpham/vielish/server/pkg/ctxbase"
)

func TestSetGetUserID(t *testing.T) {
	ctx := ctxbase.SetUserID(context.Background(), "user-abc")
	id, ok := ctxbase.GetUserID(ctx)
	if !ok {
		t.Fatal("GetUserID returned ok=false")
	}
	if id != "user-abc" {
		t.Errorf("GetUserID = %q, want %q", id, "user-abc")
	}
}

func TestGetUserID_Missing(t *testing.T) {
	_, ok := ctxbase.GetUserID(context.Background())
	if ok {
		t.Error("GetUserID should return ok=false on empty context")
	}
}

func TestSetGetRequestID(t *testing.T) {
	ctx := ctxbase.SetRequestID(context.Background(), "req-123")
	id, ok := ctxbase.GetRequestID(ctx)
	if !ok {
		t.Fatal("GetRequestID returned ok=false")
	}
	if id != "req-123" {
		t.Errorf("GetRequestID = %q, want %q", id, "req-123")
	}
}
