package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestWALEntry tests WAL entry structure
func TestWALEntry(t *testing.T) {
	entry := &WALEntry{
		Index:     1,
		Term:      1,
		Type:      WALTypeMessage,
		Data:      []byte("test data"),
		Timestamp: time.Now(),
	}

	if entry.Index != 1 {
		t.Errorf("Expected index 1, got %d", entry.Index)
	}
	if entry.Term != 1 {
		t.Errorf("Expected term 1, got %d", entry.Term)
	}
	if entry.Type != WALTypeMessage {
		t.Errorf("Expected type %s, got %s", WALTypeMessage, entry.Type)
	}
	if string(entry.Data) != "test data" {
		t.Errorf("Expected data 'test data', got %s", string(entry.Data))
	}
	if entry.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

// TestNewFileBasedWALWriter tests WAL creation
func TestNewFileBasedWALWriter(t *testing.T) {
	// Create temporary directory
	tempDir, err := ioutil.TempDir("", "wal_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		WALDir:          tempDir,
		WALSyncInterval: 100 * time.Millisecond,
	}

	// Test successful creation
	wal, err := newFileBasedWALWriter(config)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	if wal == nil {
		t.Fatal("WAL should not be nil")
	}
	defer wal.Close()

	// Verify file exists
	walFile := filepath.Join(tempDir, "wal.log")
	_, err = os.Stat(walFile)
	if err != nil {
		t.Errorf("WAL file should exist: %v", err)
	}

	// Verify initial state
	if wal.currentIndex != 0 {
		t.Errorf("Expected currentIndex 0, got %d", wal.currentIndex)
	}
	if wal.writer == nil {
		t.Error("Writer should not be nil")
	}
	if wal.stopCh == nil {
		t.Error("StopCh should not be nil")
	}
}

// TestNewFileBasedWALWriter_InvalidDir tests WAL creation with invalid directory
func TestNewFileBasedWALWriter_InvalidDir(t *testing.T) {
	config := &StorageConfig{
		WALDir:          "/invalid/path/that/does/not/exist",
		WALSyncInterval: 100 * time.Millisecond,
	}

	wal, err := newFileBasedWALWriter(config)
	if err == nil {
		t.Error("Expected error for invalid directory")
	}
	if wal != nil {
		t.Error("WAL should be nil on error")
	}
}

// TestWALAppend tests appending entries to WAL
func TestWALAppend(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "wal_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		WALDir:          tempDir,
		WALSyncInterval: 1 * time.Second,
	}

	wal, err := newFileBasedWALWriter(config)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	ctx := context.Background()

	// Test single append
	entry := &WALEntry{
		Term: 1,
		Type: WALTypeMessage,
		Data: []byte("test message"),
	}

	err = wal.Append(ctx, entry)
	if err != nil {
		t.Errorf("Failed to append entry: %v", err)
	}

	// Verify entry was assigned correct index
	if entry.Index != 1 {
		t.Errorf("Expected index 1, got %d", entry.Index)
	}
	if entry.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
	if entry.Checksum == 0 {
		t.Error("Expected non-zero checksum")
	}

	// Verify metrics
	if wal.entriesWritten != 1 {
		t.Errorf("Expected 1 entry written, got %d", wal.entriesWritten)
	}
	if wal.bytesWritten <= 0 {
		t.Errorf("Expected positive bytes written, got %d", wal.bytesWritten)
	}
}

// TestWALAppendMultiple tests appending multiple entries
func TestWALAppendMultiple(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "wal_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		WALDir:          tempDir,
		WALSyncInterval: 1 * time.Second,
	}

	wal, err := newFileBasedWALWriter(config)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	ctx := context.Background()

	// Append multiple entries
	entries := []*WALEntry{
		{Term: 1, Type: WALTypeMessage, Data: []byte("message 1")},
		{Term: 1, Type: WALTypeMessage, Data: []byte("message 2")},
		{Term: 2, Type: WALTypeMetadata, Data: []byte("metadata")},
	}

	for i, entry := range entries {
		err = wal.Append(ctx, entry)
		if err != nil {
			t.Errorf("Failed to append entry %d: %v", i, err)
		}
		expectedIndex := int64(i + 1)
		if entry.Index != expectedIndex {
			t.Errorf("Expected index %d, got %d", expectedIndex, entry.Index)
		}
	}

	// Verify final state
	if wal.entriesWritten != 3 {
		t.Errorf("Expected 3 entries written, got %d", wal.entriesWritten)
	}
	if wal.currentIndex != 3 {
		t.Errorf("Expected currentIndex 3, got %d", wal.currentIndex)
	}
}

// TestWALRead tests reading entries from WAL
func TestWALRead(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "wal_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		WALDir:          tempDir,
		WALSyncInterval: 1 * time.Second,
	}

	wal, err := newFileBasedWALWriter(config)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	ctx := context.Background()

	// Write test entries
	testData := []string{"entry1", "entry2", "entry3", "entry4", "entry5"}
	for _, data := range testData {
		entry := &WALEntry{
			Term: 1,
			Type: WALTypeMessage,
			Data: []byte(data),
		}
		err = wal.Append(ctx, entry)
		if err != nil {
			t.Fatalf("Failed to append entry: %v", err)
		}
	}

	// Force sync to ensure data is written
	err = wal.Sync(ctx)
	if err != nil {
		t.Fatalf("Failed to sync WAL: %v", err)
	}

	// Test reading all entries
	entries, err := wal.Read(ctx, 1, 10)
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(entries))
	}

	// Verify entries
	for i, entry := range entries {
		expectedIndex := int64(i + 1)
		if entry.Index != expectedIndex {
			t.Errorf("Expected index %d, got %d", expectedIndex, entry.Index)
		}
		if entry.Term != 1 {
			t.Errorf("Expected term 1, got %d", entry.Term)
		}
		if entry.Type != WALTypeMessage {
			t.Errorf("Expected type %s, got %s", WALTypeMessage, entry.Type)
		}
		if string(entry.Data) != testData[i] {
			t.Errorf("Expected data %s, got %s", testData[i], string(entry.Data))
		}
		if entry.Checksum == 0 {
			t.Error("Expected non-zero checksum")
		}
	}

	// Test reading with offset
	entries, err = wal.Read(ctx, 3, 10)
	if err != nil {
		t.Fatalf("Failed to read entries with offset: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
	if entries[0].Index != 3 {
		t.Errorf("Expected first entry index 3, got %d", entries[0].Index)
	}
	if string(entries[0].Data) != "entry3" {
		t.Errorf("Expected first entry data 'entry3', got %s", string(entries[0].Data))
	}

	// Test reading with limit
	entries, err = wal.Read(ctx, 1, 2)
	if err != nil {
		t.Fatalf("Failed to read entries with limit: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
	if entries[0].Index != 1 {
		t.Errorf("Expected first entry index 1, got %d", entries[0].Index)
	}
	if entries[1].Index != 2 {
		t.Errorf("Expected second entry index 2, got %d", entries[1].Index)
	}
}

// TestWALLatestIndex tests getting latest index
func TestWALLatestIndex(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "wal_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		WALDir:          tempDir,
		WALSyncInterval: 1 * time.Second,
	}

	wal, err := newFileBasedWALWriter(config)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	ctx := context.Background()

	// Initial index should be 0
	index, err := wal.LatestIndex(ctx)
	if err != nil {
		t.Errorf("Failed to get latest index: %v", err)
	}
	if index != 0 {
		t.Errorf("Expected initial index 0, got %d", index)
	}

	// Append entries and check index
	for i := 1; i <= 5; i++ {
		entry := &WALEntry{
			Term: 1,
			Type: WALTypeMessage,
			Data: []byte("test"),
		}
		err = wal.Append(ctx, entry)
		if err != nil {
			t.Fatalf("Failed to append entry: %v", err)
		}

		index, err = wal.LatestIndex(ctx)
		if err != nil {
			t.Errorf("Failed to get latest index: %v", err)
		}
		if index != int64(i) {
			t.Errorf("Expected index %d, got %d", i, index)
		}
	}
}

// TestWALRecovery tests WAL recovery after restart
func TestWALRecovery(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "wal_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		WALDir:          tempDir,
		WALSyncInterval: 1 * time.Second,
	}

	// Create WAL and write entries
	wal1, err := newFileBasedWALWriter(config)
	if err != nil {
		t.Fatalf("Failed to create first WAL: %v", err)
	}

	ctx := context.Background()
	testData := []string{"recovery1", "recovery2", "recovery3"}

	for _, data := range testData {
		entry := &WALEntry{
			Term: 1,
			Type: WALTypeMessage,
			Data: []byte(data),
		}
		err = wal1.Append(ctx, entry)
		if err != nil {
			t.Fatalf("Failed to append entry: %v", err)
		}
	}

	// Force sync and close
	err = wal1.Sync(ctx)
	if err != nil {
		t.Fatalf("Failed to sync WAL: %v", err)
	}
	err = wal1.Close()
	if err != nil {
		t.Fatalf("Failed to close WAL: %v", err)
	}

	// Create new WAL instance (simulating restart)
	wal2, err := newFileBasedWALWriter(config)
	if err != nil {
		t.Fatalf("Failed to create second WAL: %v", err)
	}
	defer wal2.Close()

	// Verify recovery
	index, err := wal2.LatestIndex(ctx)
	if err != nil {
		t.Errorf("Failed to get latest index: %v", err)
	}
	if index != 3 {
		t.Errorf("Expected recovered index 3, got %d", index)
	}

	// Verify we can read recovered entries
	entries, err := wal2.Read(ctx, 1, 10)
	if err != nil {
		t.Fatalf("Failed to read recovered entries: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 recovered entries, got %d", len(entries))
	}

	for i, entry := range entries {
		expectedIndex := int64(i + 1)
		if entry.Index != expectedIndex {
			t.Errorf("Expected index %d, got %d", expectedIndex, entry.Index)
		}
		if string(entry.Data) != testData[i] {
			t.Errorf("Expected data %s, got %s", testData[i], string(entry.Data))
		}
	}

	// Verify we can continue writing
	newEntry := &WALEntry{
		Term: 2,
		Type: WALTypeMessage,
		Data: []byte("post-recovery"),
	}
	err = wal2.Append(ctx, newEntry)
	if err != nil {
		t.Errorf("Failed to append post-recovery entry: %v", err)
	}
	if newEntry.Index != 4 {
		t.Errorf("Expected post-recovery index 4, got %d", newEntry.Index)
	}
}

// TestWALSync tests manual sync
func TestWALSync(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "wal_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		WALDir:          tempDir,
		WALSyncInterval: 10 * time.Second, // Long interval
	}

	wal, err := newFileBasedWALWriter(config)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	ctx := context.Background()

	// Append entry
	entry := &WALEntry{
		Term: 1,
		Type: WALTypeMessage,
		Data: []byte("sync test"),
	}
	err = wal.Append(ctx, entry)
	if err != nil {
		t.Fatalf("Failed to append entry: %v", err)
	}

	// Manual sync
	err = wal.Sync(ctx)
	if err != nil {
		t.Errorf("Failed to sync: %v", err)
	}
	if wal.syncCount <= 0 {
		t.Error("Expected sync count to be greater than 0")
	}
}

// TestWALEmptyRead tests reading from empty WAL
func TestWALEmptyRead(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "wal_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		WALDir:          tempDir,
		WALSyncInterval: 1 * time.Second,
	}

	wal, err := newFileBasedWALWriter(config)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	ctx := context.Background()

	// Read from empty WAL
	entries, err := wal.Read(ctx, 1, 10)
	if err != nil {
		t.Errorf("Failed to read from empty WAL: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries from empty WAL, got %d", len(entries))
	}

	// Latest index should be 0
	index, err := wal.LatestIndex(ctx)
	if err != nil {
		t.Errorf("Failed to get latest index: %v", err)
	}
	if index != 0 {
		t.Errorf("Expected latest index 0 for empty WAL, got %d", index)
	}
}

// TestWALConcurrency tests concurrent operations
func TestWALConcurrency(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "wal_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		WALDir:          tempDir,
		WALSyncInterval: 100 * time.Millisecond,
	}

	wal, err := newFileBasedWALWriter(config)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	ctx := context.Background()
	numGoroutines := 10
	entriesPerGoroutine := 10

	// Concurrent writes
	done := make(chan struct{}, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- struct{}{} }()
			for j := 0; j < entriesPerGoroutine; j++ {
				entry := &WALEntry{
					Term: int64(id),
					Type: WALTypeMessage,
					Data: []byte(fmt.Sprintf("goroutine-%d-entry-%d", id, j)),
				}
				err := wal.Append(ctx, entry)
				if err != nil {
					t.Errorf("Failed to append entry: %v", err)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all entries were written
	index, err := wal.LatestIndex(ctx)
	if err != nil {
		t.Errorf("Failed to get latest index: %v", err)
	}
	expectedTotal := int64(numGoroutines * entriesPerGoroutine)
	if index != expectedTotal {
		t.Errorf("Expected latest index %d, got %d", expectedTotal, index)
	}

	// Force sync and read all entries
	err = wal.Sync(ctx)
	if err != nil {
		t.Fatalf("Failed to sync WAL: %v", err)
	}

	entries, err := wal.Read(ctx, 1, int32(numGoroutines*entriesPerGoroutine))
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}
	if len(entries) != numGoroutines*entriesPerGoroutine {
		t.Errorf("Expected %d entries, got %d", numGoroutines*entriesPerGoroutine, len(entries))
	}

	// Verify sequential indexes
	for i, entry := range entries {
		expectedIndex := int64(i + 1)
		if entry.Index != expectedIndex {
			t.Errorf("Expected index %d, got %d", expectedIndex, entry.Index)
		}
	}
}

// Benchmark tests
func BenchmarkWALAppend(b *testing.B) {
	tempDir, err := ioutil.TempDir("", "wal_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		WALDir:          tempDir,
		WALSyncInterval: 1 * time.Second,
	}

	wal, err := newFileBasedWALWriter(config)
	if err != nil {
		b.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	ctx := context.Background()
	entry := &WALEntry{
		Term: 1,
		Type: WALTypeMessage,
		Data: make([]byte, 1024), // 1KB entry
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := wal.Append(ctx, entry)
		if err != nil {
			b.Fatalf("Failed to append entry: %v", err)
		}
	}
}

func BenchmarkWALRead(b *testing.B) {
	tempDir, err := ioutil.TempDir("", "wal_bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &StorageConfig{
		WALDir:          tempDir,
		WALSyncInterval: 1 * time.Second,
	}

	wal, err := newFileBasedWALWriter(config)
	if err != nil {
		b.Fatalf("Failed to create WAL: %v", err)
	}
	defer wal.Close()

	ctx := context.Background()

	// Pre-populate with entries
	for i := 0; i < 1000; i++ {
		entry := &WALEntry{
			Term: 1,
			Type: WALTypeMessage,
			Data: make([]byte, 1024),
		}
		err := wal.Append(ctx, entry)
		if err != nil {
			b.Fatalf("Failed to append entry: %v", err)
		}
	}

	err = wal.Sync(ctx)
	if err != nil {
		b.Fatalf("Failed to sync WAL: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := wal.Read(ctx, 1, 100)
		if err != nil {
			b.Fatalf("Failed to read entries: %v", err)
		}
	}
}
