package sse

import "net/url"

// Subscriber ...
type Subscriber struct {
	quit       chan *Subscriber
	connection chan *Event
	removed    chan struct{}
	eventid    int
	URL        *url.URL
}

// Close will let the stream know that the clients connection has terminated
func (s *Subscriber) close() {
	s.quit <- s
	if s.removed != nil {
		<-s.removed
	}
}
