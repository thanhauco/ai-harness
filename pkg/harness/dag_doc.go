// Package harness includes the DAG pipeline execution engine.
//
// Architecture Overview:
//
//	[Step A] ------+
//	               |---> [Step C] ---> [Step D]
//	[Step B] ------+
//
// Steps within the same dependency tier execute concurrently, bounded by
// the configurable worker concurrency limit.
package harness
