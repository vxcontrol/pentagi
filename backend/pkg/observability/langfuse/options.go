package langfuse

import "time"

type ObserverOption func(*observer)

func WithProject(project string) ObserverOption {
	return func(o *observer) {
		o.project = project
	}
}

func WithRelease(release string) ObserverOption {
	return func(o *observer) {
		o.release = release
	}
}

func WithSendInterval(interval time.Duration) ObserverOption {
	return func(o *observer) {
		o.interval = interval
	}
}

func WithSendTimeout(timeout time.Duration) ObserverOption {
	return func(o *observer) {
		o.timeout = timeout
	}
}

func WithQueueSize(size int) ObserverOption {
	return func(o *observer) {
		o.queueSize = size
	}
}
