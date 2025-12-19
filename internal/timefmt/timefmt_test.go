package timefmt

import (
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	now := time.Now()
	result := Format(now, "2006-01-02")
	if result == "" {
		t.Error("Format returned empty string")
	}
}

