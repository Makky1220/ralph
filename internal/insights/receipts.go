package insights

import (
	"bufio"
	"encoding/json"
	"os"
)

// Receipt represents one model-commanding receipt from the org runtime.
type Receipt struct {
	TS                     string `json:"ts"`
	OrgID                  string `json:"org_id"`
	SeatID                 string `json:"seat_id"`
	Role                   string `json:"role,omitempty"`
	Driver                 string `json:"driver"`
	CommandedModel         string `json:"commanded_model"`
	ReportedEffectiveModel string `json:"reported_effective_model,omitempty"`
	Honored                string `json:"honored"` // "true" | "false" | "unknown"
	Reason                 string `json:"reason,omitempty"`
}

// Honored constants for tri-state model-routing honored field.
const (
	HonoredTrue    = "true"
	HonoredFalse   = "false"
	HonoredUnknown = "unknown"
)

// ReceiptStats holds counters from a ReadReceipts call.
type ReceiptStats struct {
	LinesRead    int
	SkippedLines int
}

// ReadReceipts reads receipts from a JSONL file at path.
// A missing file returns empty, no error (graceful degradation).
// Corrupt lines are counted in stats.SkippedLines and skipped (non-fatal).
func ReadReceipts(path string) ([]Receipt, ReceiptStats, error) {
	var stats ReceiptStats

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, stats, nil
		}
		return nil, stats, err
	}
	defer func() { _ = f.Close() }()

	var receipts []Receipt
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		stats.LinesRead++

		var r Receipt
		if err := json.Unmarshal(line, &r); err != nil {
			stats.SkippedLines++
			continue
		}
		receipts = append(receipts, r)
	}

	if err := scanner.Err(); err != nil {
		return nil, stats, err
	}

	return receipts, stats, nil
}
