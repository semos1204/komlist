package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/task"
)

func TestBlock_AddsDependency(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	a, _ := svc.Add(ctx, "deploy")
	b, _ := svc.Add(ctx, "tests")

	got, err := svc.Block(ctx, a.ID, b.ID)
	if err != nil {
		t.Fatalf("block: %v", err)
	}
	if len(got.DependsOn) != 1 || got.DependsOn[0] != b.ID {
		t.Errorf("DependsOn = %v, want [%d]", got.DependsOn, b.ID)
	}

	// idempotent
	got, _ = svc.Block(ctx, a.ID, b.ID)
	if len(got.DependsOn) != 1 {
		t.Errorf("duplicate block should be a no-op, got %v", got.DependsOn)
	}
}

func TestBlock_SelfDependency(t *testing.T) {
	svc, _ := newSvc()
	a, _ := svc.Add(context.Background(), "x")
	if _, err := svc.Block(context.Background(), a.ID, a.ID); !errors.Is(err, service.ErrSelfDependency) {
		t.Errorf("got %v, want ErrSelfDependency", err)
	}
}

func TestBlock_CycleRejected(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	a, _ := svc.Add(ctx, "a")
	b, _ := svc.Add(ctx, "b")
	if _, err := svc.Block(ctx, a.ID, b.ID); err != nil { // a depends on b
		t.Fatalf("block a->b: %v", err)
	}
	// b depends on a would close a cycle
	if _, err := svc.Block(ctx, b.ID, a.ID); !errors.Is(err, service.ErrDependencyCycle) {
		t.Errorf("got %v, want ErrDependencyCycle", err)
	}
}

func TestBlock_TransitiveCycleRejected(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	a, _ := svc.Add(ctx, "a")
	b, _ := svc.Add(ctx, "b")
	c, _ := svc.Add(ctx, "c")
	_, _ = svc.Block(ctx, a.ID, b.ID) // a -> b
	_, _ = svc.Block(ctx, b.ID, c.ID) // b -> c
	// c -> a would create a->b->c->a cycle
	if _, err := svc.Block(ctx, c.ID, a.ID); !errors.Is(err, service.ErrDependencyCycle) {
		t.Errorf("got %v, want ErrDependencyCycle", err)
	}
}

func TestUnblock(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	a, _ := svc.Add(ctx, "a")
	b, _ := svc.Add(ctx, "b")
	_, _ = svc.Block(ctx, a.ID, b.ID)
	got, err := svc.Unblock(ctx, a.ID, b.ID)
	if err != nil {
		t.Fatalf("unblock: %v", err)
	}
	if len(got.DependsOn) != 0 {
		t.Errorf("DependsOn = %v, want empty", got.DependsOn)
	}
}

func TestBlockedSet(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	a, _ := svc.Add(ctx, "deploy")
	b, _ := svc.Add(ctx, "tests")
	_, _ = svc.Block(ctx, a.ID, b.ID)

	blocked, err := svc.BlockedSet(ctx)
	if err != nil {
		t.Fatalf("blockedSet: %v", err)
	}
	if !blocked[a.ID] {
		t.Error("deploy should be blocked while tests is not done")
	}

	// completing the blocker unblocks
	if _, err := svc.ChangeStatus(ctx, b.ID, task.StatusDone); err != nil {
		t.Fatalf("done: %v", err)
	}
	blocked, _ = svc.BlockedSet(ctx)
	if blocked[a.ID] {
		t.Error("deploy should be unblocked once tests is done")
	}
}
