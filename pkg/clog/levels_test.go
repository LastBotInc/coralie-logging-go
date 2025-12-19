package clog

import "testing"

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level  Level
		expect string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelSuccess, "SUCCESS"},
		{LevelWarning, "WARNING"},
		{LevelFail, "FAIL"},
		{LevelError, "ERROR"},
		{LevelCatastrophe, "CATASTROPHE"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expect {
			t.Errorf("Level.String() = %v, want %v", got, tt.expect)
		}
	}
}

