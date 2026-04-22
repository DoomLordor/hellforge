package locker

import "context"

// Locker provide distributed lock interface
type Locker interface {
	Lock(ctx context.Context, task *LockKey) error
}

// LockKey struct for lock key description with start/stop callbacks if needed
type LockKey struct {
	Key     string
	OnStart func() error
	OnStop  func() error
}
