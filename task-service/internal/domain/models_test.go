package domain

import "testing"

func TestTaskTransitions(t *testing.T) {
	if !CanTransition(StatusTodo, StatusInProgress) {
		t.Fatal("todo should transition to in progress")
	}
	if CanTransition(StatusBacklog, StatusDone) {
		t.Fatal("backlog must not transition directly to done")
	}
}

func TestPositionBetween(t *testing.T) {
	a, b := 100.0, 200.0
	if got := PositionBetween(&a, &b); got != 150 {
		t.Fatalf("got %v", got)
	}
	if got := PositionBetween(nil, nil); got != 1024 {
		t.Fatalf("got %v", got)
	}
}
