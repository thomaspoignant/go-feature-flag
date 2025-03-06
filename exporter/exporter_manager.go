package exporter

type Manager[T any] interface {
	AddEvent(event T)
	StartDaemon()
	Close()
}
