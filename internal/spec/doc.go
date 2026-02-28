// Package spec provides task registry and change management for OpenSpec workflows.
// It includes domain types for changes, tasks, and dependencies, plus a file-based
// store for OpenSpec project structure.
//
// The package separates concerns:
// - Domain types: ChangeMetadata, Task, DependencyInfo, RuntimeState, SyncMeta
// - Store: File-based operations for specs and changes
// - storage subpackage: BoltDB implementation for task caching
package spec
