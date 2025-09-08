package storage

import "time"

type SegmentStore interface {
	// Create a new segment
	CreateSegment(topic string, partition int32, baseOffset int64) (Segment, error)

	// Open an existing segment
	OpenSegment(topic string, partition int32, baseOffset int64) (Segment, error)

	// List all segments for a topic partition
	ListSegments(topic string, partition int32) ([]SegmentInfo, error)

	// Delete a segment
	DeleteSegment(topic string, partition int32, baseOffset int64) error

	// Close all segments
	Close() error
}

// SegmentInfo contains metadata about a segment
type SegmentInfo struct {
	Topic        string    `json:"topic"`
	Partition    int32     `json:"partition"`
	BaseOffset   int64     `json:"base_offset"`
	LatestOffset int64     `json:"latest_offset"`
	Size         int64     `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}