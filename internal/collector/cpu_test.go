package collector

import (
	"testing"
	"time"
)

func TestCPUExport_DeepCopy(t *testing.T) {
	original := CPUExport{
		History: []CPUSample{
			{
				Timestamp: time.Unix(100, 0),
				Data: CPUData{
					Usage: CPUUsage{
						Valid: true,
						Data:  25,
					},
					Cores: CoresStat{
						Valid: true,
						Data: []CoreStat{
							{ID: 0, Usage: 10},
							{ID: 1, Usage: 20},
						},
					},
				},
			},
		},
	}

	copyExport := original.DeepCopy()

	if len(copyExport.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(copyExport.History))
	}

	copyExport.History[0].Data.Cores.Data[0].Usage = 99

	if got := original.History[0].Data.Cores.Data[0].Usage; got != 10 {
		t.Errorf("original core usage = %v, want 10", got)
	}
}

func TestCPUExport_DeepCopy_EmptyHistory(t *testing.T) {
	original := CPUExport{}

	copyExport := original.DeepCopy()

	if len(copyExport.History) != 0 {
		t.Errorf("history length = %d, want 0", len(copyExport.History))
	}
}
