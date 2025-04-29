// Package models defines the core data structures used throughout the application.
package models

// Metric represents a system metric that can be either a gauge or counter type.
type Metric struct {
	MType string
	Name  string
	Value float64
	Delta int64
}
