// Package clog: hooks system for custom log processing.
package clog

// Hook is an interface for custom log processing.
type Hook interface {
	OnLog(Event)
}

