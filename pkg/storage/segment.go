package storage

import "time"

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

