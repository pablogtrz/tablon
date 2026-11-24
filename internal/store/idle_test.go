package store

import (
	"context"
	"testing"
)

func TestEnterAndExitIdle(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	ctx := context.Background()
	if err := st.EnterIdle(ctx); err != nil {
		t.Fatalf("enter idle: %v", err)
	}
	if err := st.Ping(ctx); err != nil {
		t.Fatalf("ping after idle: %v", err)
	}
	st.ExitIdle()
	if err := st.Ping(ctx); err != nil {
		t.Fatalf("ping after exit idle: %v", err)
	}
}
