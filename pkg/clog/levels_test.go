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

func TestLevel_AtLeast(t *testing.T) {
	tests := []struct {
		level Level
		min   Level
		want  bool
	}{
		{LevelDebug, LevelDebug, true},
		{LevelInfo, LevelDebug, true},
		{LevelCatastrophe, LevelDebug, true},
		{LevelDebug, LevelInfo, false},
		{LevelWarning, LevelWarning, true},
		{LevelError, LevelWarning, true},
		{LevelInfo, LevelCatastrophe, false},
	}
	for _, tt := range tests {
		if got := tt.level.AtLeast(tt.min); got != tt.want {
			t.Errorf("Level(%s).AtLeast(%s) = %v, want %v", tt.level.String(), tt.min.String(), got, tt.want)
		}
	}
}
