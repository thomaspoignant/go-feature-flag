package notifier

// Notifier is the interface to represent a GO Feature Flag notifier
type Notifier interface {
	// Notify is the function doing all the work when a Notifier is called.
	Notify(cache DiffCache) error
}
