package main

import (
	"io"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}
		c := &Client{conn, make(chan *[]byte, 1024)}

		go func() {
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					c.Close()
				}
				m := string(message)
				if m == "segment" {
					b := <-c.queue
					cw, err := c.NextWriter(websocket.BinaryMessage)
					if err != nil {
						return
					}
					_, err = cw.Write(*b)
					if err != nil {
						return
					}
					err = cw.Close()
					if err != nil {
						return
					}
				}
			}
		}()

		go func() {
			var cw io.Writer = c
			s := newStream("./test.webm")
			s.writers[&cw] = make(chan bool)
			go s.run()
			<-s.writers[&cw]
		}()
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./index.html")
	})
	if err := http.ListenAndServe(":8000", nil); err != nil {
		panic(err)
	}
}

type Client struct {
	*websocket.Conn
	queue chan *[]byte
}

func (c *Client) Write(b []byte) (int, error) {
	c.queue <- &b
	return 0, nil
}

type Stream struct {
	path    string
	writers map[*io.Writer]chan bool
	stdout  chan *[]byte
}

func newStream(path string) *Stream {
	return &Stream{
		path:    path,
		writers: make(map[*io.Writer]chan bool, 256),
		stdout:  make(chan *[]byte, 1024),
	}
}

func (s *Stream) run() {
	go s.write()
	size := 1024 * 50
	f, err := os.Open(s.path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for {
		if len(s.writers) == 0 {
			close(s.stdout)
			return
		}
		buf := make([]byte, size)
		_, err := io.ReadFull(f, buf)
		if err != nil {
			close(s.stdout)
			return
		}
		s.stdout <- &buf
	}
}

func (s *Stream) write() {
	for buf := range s.stdout {
		for w, ch := range s.writers {
			_, err := (*w).Write(*buf)
			if err != nil {
				delete(s.writers, w)
				close(ch)
			}
		}
	}
}
