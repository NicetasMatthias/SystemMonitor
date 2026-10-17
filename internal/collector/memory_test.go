package collector_test

import (
	"testing"
	"time"

	"github.com/NicetasMatthias/SystemMonitor/internal/collector"
)

func TestMemoryExport_DeepCopy(t *testing.T) {
	original := collector.MemoryExport{
		History: []collector.MemorySample{
			{
				Valid:     true,
				Timestamp: time.Unix(100, 0),
				Total:     1000,
				Used:      400,
				Available: 600,
			},
		},
	}

	copyExport := original.DeepCopy()

	if len(copyExport.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(copyExport.History))
	}

	copyExport.History[0].Used = 999

	if got := original.History[0].Used; got != 400 {
		t.Errorf("original Used = %d, want 400", got)
	}
}

func TestMemoryExport_DeepCopy_EmptyHistory(t *testing.T) {
	original := collector.MemoryExport{}

	copyExport := original.DeepCopy()

	if len(copyExport.History) != 0 {
		t.Errorf("history length = %d, want 0", len(copyExport.History))
	}
}
