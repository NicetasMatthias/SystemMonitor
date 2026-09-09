package logger

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type LogEntry struct {
	Time    time.Time      `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"message"`
	Attrs   map[string]any `json:"attrs,omitempty"`
}

type LogSubscription struct {
	History []LogEntry
	Events  <-chan LogEntry
	Cancel  func()
}

type LogStore interface {
	Subscribe() LogSubscription
}

var (
	storageState *storageHandlerState
	storageOnce  sync.Once
)

type storageHandlerState struct {
	mx sync.RWMutex

	entries []LogEntry
	maxSize int

	subscribers map[chan LogEntry]struct{}
}

type storageHandler struct {
	state *storageHandlerState

	attrs  []slog.Attr
	groups []string
}

func newStorageHandler() *storageHandler {

	storageOnce.Do(func() {
		storageState = &storageHandlerState{
			maxSize:     1000,
			subscribers: make(map[chan LogEntry]struct{}),
		}
	})

	return &storageHandler{
		state: storageState,
	}

}

func Logs() LogStore {
	return storageState
}

func (h *storageHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *storageHandler) Handle(_ context.Context, r slog.Record) error {

	entry := LogEntry{
		Time:    r.Time,
		Level:   r.Level.String(),
		Message: r.Message,
		Attrs:   make(map[string]any),
	}

	for _, attrs := range h.attrs {
		h.addAttr(entry.Attrs, h.groups, attrs)
	}

	r.Attrs(func(attr slog.Attr) bool {
		entry.Attrs[attr.Key] = attr.Value.Any()
		return true
	})

	h.state.process(entry)
	return nil
}

func (h *storageHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	newAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)

	return &storageHandler{
		state:  h.state,
		attrs:  newAttrs,
		groups: append([]string(nil), h.groups...),
	}
}

func (h *storageHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	newGroups := make([]string, 0, len(h.groups)+1)
	newGroups = append(newGroups, h.groups...)
	newGroups = append(newGroups, name)

	return &storageHandler{
		state:  h.state,
		attrs:  append([]slog.Attr(nil), h.attrs...),
		groups: newGroups,
	}
}

func (s *storageHandlerState) Entries() []LogEntry {
	s.mx.RLock()
	defer s.mx.RUnlock()

	entries := make([]LogEntry, len(s.entries))
	copy(entries, s.entries)
	return entries
}

func (s *storageHandlerState) Subscribe() LogSubscription {
	ch := make(chan LogEntry, s.maxSize)

	s.mx.Lock()
	defer s.mx.Unlock()
	s.subscribers[ch] = struct{}{}

	cancel := func() {
		s.mx.Lock()
		delete(s.subscribers, ch)
		s.mx.Unlock()
	}

	history := make([]LogEntry, len(s.entries))
	copy(history, s.entries)

	return LogSubscription{
		History: history,
		Events:  ch,
		Cancel:  cancel,
	}
}

func (h *storageHandler) addAttr(
	dst map[string]any,
	groups []string,
	attr slog.Attr,
) {
	attr.Value = attr.Value.Resolve()

	if attr.Equal(slog.Attr{}) {
		return
	}

	value := attr.Value.Any()

	if len(groups) == 0 {
		dst[attr.Key] = value
		return
	}

	current := dst

	for _, group := range groups {
		value, ok := current[group]

		if !ok {
			next := make(map[string]any)
			current[group] = next
			current = next
			continue
		}

		next, ok := value.(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[group] = next
		}

		current = next
	}

	current[attr.Key] = value
}

func (s *storageHandlerState) process(e LogEntry) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.entries = append(s.entries, e)
	if len(s.entries) > s.maxSize {
		s.entries = s.entries[len(s.entries)-s.maxSize:]
	}

	for ch := range s.subscribers {
		select {
		case ch <- e:
		default:
		}
	}
}
