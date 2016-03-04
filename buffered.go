package logstreamer

import (
	"sync"
	"time"
)

// LogEntry represents a log entry. The log number indicates the log
// number since the start of the stream. If an observer observes a
// gap in log numbers, then they have lost log lines. This is caused
// by the observer not being able to process logs as fast as they're
// being generated.
type LogEntry struct {
	Number    int64
	Timestamp time.Time
	Line      string
}

// BufferedLogStream is a buffered stream of logs. It allows observers
// to view X amount of lines in the past upon creation.
type BufferedLogStream struct {
	historyMut sync.Mutex
	head       int
	len        int
	totalPos   int64
	history    []LogEntry

	observersMut   sync.Mutex
	nextObserverID int
	observers      map[int]chan<- LogEntry
}

// NewBufferedLogStream creates a new BufferedLogStream, with maxLines
// as the buffer size (in log entries).
func NewBufferedLogStream(maxLines int) *BufferedLogStream {
	return &BufferedLogStream{
		history:   make([]LogEntry, maxLines),
		observers: make(map[int]chan<- LogEntry),
	}
}

// WriteLine writes a line to the stream. The line is timestamped based
// on insertion time, not observed time.
func (b *BufferedLogStream) WriteLine(line string) error {
	entry := LogEntry{
		Timestamp: time.Now(),
		Line:      line,
	}

	b.historyMut.Lock()
	defer b.historyMut.Unlock()

	entry.Number = b.totalPos

	b.history[b.head] = entry

	b.observersMut.Lock()
	for _, out := range b.observers {
		select {
		case out <- entry:
		default:
		}
	}
	b.observersMut.Unlock()

	b.head = (b.head + 1) % len(b.history)
	b.totalPos++

	if b.len < len(b.history) {
		b.len++
	}

	return nil
}

// NewObserver creates a new StreamObserver, pre-populated with whatever
// is currently in the buffer. Callers should call Close() when finished.
func (b *BufferedLogStream) NewObserver() StreamObserver {
	obs := StreamObserver{
		stream:       b,
		observerChan: make(chan LogEntry, len(b.history)),
	}

	b.observersMut.Lock()

	obs.observerID = b.nextObserverID
	b.observers[obs.observerID] = obs.observerChan
	b.nextObserverID++

	b.observersMut.Unlock()

	b.historyMut.Lock()
	start := (b.head - b.len) % len(b.history)
	if start < 0 {
		start += len(b.history)
	}

	for i := 0; i < b.len; i++ {
		obs.observerChan <- b.history[start]
		start = (start + 1) % len(b.history)
	}

	b.historyMut.Unlock()

	return obs
}

// StreamObserver is an observer to a stream. Logs are sent through
// the streams channel. Users of StreamObserver should call Close()
// when finished with the observer.
type StreamObserver struct {
	stream *BufferedLogStream

	observerID   int
	observerChan chan LogEntry
}

// Chan returns a receive-only channel of LogEntries.
func (s *StreamObserver) Chan() <-chan LogEntry {
	return s.observerChan
}

// Close closes the stream observer, freeing any resources
// required by the associated stream.
func (s *StreamObserver) Close() error {
	s.stream.observersMut.Lock()
	delete(s.stream.observers, s.observerID)
	s.stream.observersMut.Unlock()

	close(s.observerChan)

	return nil
}
