package snowflake

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	nodeIDBits     = 10
	sequenceBits   = 12
	maxNodeID      = int64(-1) ^ (int64(-1) << nodeIDBits)
	maxSequence    = int64(-1) ^ (int64(-1) << sequenceBits)
	nodeIDShift    = sequenceBits
	timestampShift = sequenceBits + nodeIDBits
)

var (
	DefaultEpoch            = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	ErrClockMovedBack       = errors.New("clock moved backwards")
	ErrInvalidNodeID        = fmt.Errorf("node id must be between 0 and %d", maxNodeID)
	ErrTimestampBeforeEpoch = errors.New("timestamp is before snowflake epoch")
	ErrTimestampTooHigh     = errors.New("timestamp exceeds snowflake limit")
)

type Generator struct {
	mu            sync.Mutex
	nodeID        int64
	epoch         time.Time
	lastTimestamp int64
	sequence      int64
	now           func() time.Time
}

type Parts struct {
	Timestamp time.Time
	NodeID    int64
	Sequence  int64
}

func New(nodeID int64) (*Generator, error) {
	return NewWithEpoch(nodeID, DefaultEpoch)
}

func NewWithEpoch(nodeID int64, epoch time.Time) (*Generator, error) {
	if nodeID < 0 || nodeID > maxNodeID {
		return nil, ErrInvalidNodeID
	}
	return &Generator{
		nodeID:        nodeID,
		epoch:         epoch.UTC(),
		lastTimestamp: -1,
		now:           time.Now,
	}, nil
}

func (g *Generator) NextID() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := g.currentTimestamp()
	if timestamp < 0 {
		return 0, fmt.Errorf("%w: current=%d", ErrTimestampBeforeEpoch, timestamp)
	}
	if timestamp < g.lastTimestamp {
		return 0, fmt.Errorf("%w: last=%d current=%d", ErrClockMovedBack, g.lastTimestamp, timestamp)
	}

	if timestamp == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			timestamp = g.waitNextMillis(timestamp)
		}
	} else {
		g.sequence = 0
	}

	if timestamp >= 1<<41 {
		return 0, ErrTimestampTooHigh
	}

	g.lastTimestamp = timestamp
	return (timestamp << timestampShift) | (g.nodeID << nodeIDShift) | g.sequence, nil
}

func (g *Generator) Decompose(id int64) Parts {
	timestamp := id >> timestampShift
	nodeID := (id >> nodeIDShift) & maxNodeID
	sequence := id & maxSequence

	return Parts{
		Timestamp: g.epoch.Add(time.Duration(timestamp) * time.Millisecond),
		NodeID:    nodeID,
		Sequence:  sequence,
	}
}

func (g *Generator) currentTimestamp() int64 {
	return g.now().UTC().Sub(g.epoch).Milliseconds()
}

func (g *Generator) waitNextMillis(timestamp int64) int64 {
	for timestamp <= g.lastTimestamp {
		time.Sleep(time.Millisecond)
		timestamp = g.currentTimestamp()
	}
	return timestamp
}
