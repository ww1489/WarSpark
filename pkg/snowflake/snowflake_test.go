package snowflake

import (
	"errors"
	"testing"
	"time"
)

func TestNextIDIsIncreasing(t *testing.T) {
	generator, err := New(12)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	first, err := generator.NextID()
	if err != nil {
		t.Fatalf("NextID returned error: %v", err)
	}
	second, err := generator.NextID()
	if err != nil {
		t.Fatalf("NextID returned error: %v", err)
	}
	if second <= first {
		t.Fatalf("expected second id to be greater than first: first=%d second=%d", first, second)
	}
}

func TestDecompose(t *testing.T) {
	epoch := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	generator, err := NewWithEpoch(7, epoch)
	if err != nil {
		t.Fatalf("NewWithEpoch returned error: %v", err)
	}
	generator.now = func() time.Time {
		return epoch.Add(1234 * time.Millisecond)
	}

	id, err := generator.NextID()
	if err != nil {
		t.Fatalf("NextID returned error: %v", err)
	}

	parts := generator.Decompose(id)
	if !parts.Timestamp.Equal(epoch.Add(1234 * time.Millisecond)) {
		t.Fatalf("unexpected timestamp: %s", parts.Timestamp)
	}
	if parts.NodeID != 7 {
		t.Fatalf("unexpected node id: %d", parts.NodeID)
	}
	if parts.Sequence != 0 {
		t.Fatalf("unexpected sequence: %d", parts.Sequence)
	}
}

func TestRejectsInvalidNodeID(t *testing.T) {
	_, err := New(maxNodeID + 1)
	if !errors.Is(err, ErrInvalidNodeID) {
		t.Fatalf("expected ErrInvalidNodeID, got %v", err)
	}
}

func TestRejectsClockMovedBackwards(t *testing.T) {
	epoch := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	generator, err := NewWithEpoch(1, epoch)
	if err != nil {
		t.Fatalf("NewWithEpoch returned error: %v", err)
	}

	times := []time.Time{
		epoch.Add(2 * time.Millisecond),
		epoch.Add(time.Millisecond),
	}
	index := 0
	generator.now = func() time.Time {
		value := times[index]
		index++
		return value
	}

	if _, err := generator.NextID(); err != nil {
		t.Fatalf("first NextID returned error: %v", err)
	}
	_, err = generator.NextID()
	if !errors.Is(err, ErrClockMovedBack) {
		t.Fatalf("expected ErrClockMovedBack, got %v", err)
	}
}

func TestRejectsTimestampBeforeEpoch(t *testing.T) {
	epoch := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	generator, err := NewWithEpoch(1, epoch)
	if err != nil {
		t.Fatalf("NewWithEpoch returned error: %v", err)
	}
	generator.now = func() time.Time {
		return epoch.Add(-time.Millisecond)
	}

	_, err = generator.NextID()
	if !errors.Is(err, ErrTimestampBeforeEpoch) {
		t.Fatalf("expected ErrTimestampBeforeEpoch, got %v", err)
	}
}
