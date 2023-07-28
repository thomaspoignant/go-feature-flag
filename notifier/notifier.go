package notifier

type Notifier interface {
	Notify(cache DiffCache) error
}
