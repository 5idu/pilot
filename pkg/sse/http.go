package sse

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ServeHTTP serves new connections with events for a given stream ...
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, err := w.(http.Flusher)
	if !err {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for k, v := range s.Headers {
		w.Header().Set(k, v)
	}

	// Get the StreamID from the URL
	streamID := r.URL.Query().Get("type")
	if streamID == "" {
		http.Error(w, "Please specify a stream!", http.StatusInternalServerError)
		return
	}

	stream := s.getStream(streamID)

	if stream == nil {
		if !s.AutoStream {
			http.Error(w, "Stream not found!", http.StatusInternalServerError)
			return
		}

		stream = s.CreateStream(streamID)
	}

	// TODO: 如果与服务器端的连接中断，当浏览器端再次进行连接时，会通过 HTTP 头“Last-Event-ID”来声明最后一次接收到的事件的标识符。服务器端可以通过浏览器端发送的事件标识符来确定从哪个事件开始来继续连接。
	eventid := 0
	if id := r.Header.Get("Last-Event-ID"); id != "" {
		var err error
		eventid, err = strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Last-Event-ID must be a number!", http.StatusBadRequest)
			return
		}
	}

	// Create the stream subscriber
	sub := stream.addSubscriber(eventid, r.URL)

	go func() {
		<-r.Context().Done()

		sub.close()

		if s.AutoStream && !s.AutoReplay && stream.getSubscriberCount() == 0 {
			s.RemoveStream(streamID)
		}
	}()

	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Push events to client
	for ev := range sub.connection {
		// If the data buffer is an empty string abort.
		if len(ev.Data) == 0 && len(ev.Comment) == 0 {
			break
		}

		// if the event has expired, dont send it
		if s.EventTTL != 0 && time.Now().After(ev.timestamp.Add(s.EventTTL)) {
			continue
		}

		if len(ev.Data) > 0 {
			fmt.Fprintf(w, "id: %s\n", ev.ID)
			if s.SplitData {
				sd := bytes.Split(ev.Data, []byte("\n"))
				for i := range sd {
					fmt.Fprintf(w, "data: %s\n", sd[i])
				}
			} else {
				if bytes.HasPrefix(ev.Data, []byte(":")) {
					fmt.Fprintf(w, "%s\n", ev.Data)
				} else {
					fmt.Fprintf(w, "data: %s\n", ev.Data)
				}
			}

			if len(ev.Event) > 0 {
				fmt.Fprintf(w, "event: %s\n", ev.Event)
			}

			if len(ev.Retry) > 0 {
				fmt.Fprintf(w, "retry: %s\n", ev.Retry)
			}
		}

		if len(ev.Comment) > 0 {
			fmt.Fprintf(w, ": %s\n", ev.Comment)
		}

		fmt.Fprint(w, "\n")

		flusher.Flush()
	}
}