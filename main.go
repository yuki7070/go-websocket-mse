package main

import (
	"io"
	"net/http"
	"os"
	"fmt"
	//"os/exec"
	"time"

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

	f, err := os.Open("test.webm")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	/*
	recdvb := exec.Command("recdvb", []string{"--b25", "--strip", "21", "-", "-"}...)
	pipe, err := recdvb.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}
	recdvb.Stderr = os.Stderr
	ffmpeg := exec.Command("/usr/bin/ffmpeg", []string{"-i", "pipe:0", "-threads", "0", "-s", "720x480", "-c:v", "vp9", "-f", "webm", "-deadline", "realtime", "-speed", "4", "-cpu-used", "-8", "pipe:1"}...)
	stdout, err := ffmpeg.StdoutPipe()
	ffmpeg.Stdin = pipe
	if err != nil {
		fmt.Println(err)
	}
	ffmpeg.Stderr = os.Stderr
	*/

	stdout, err := os.Open("./test.webm")
	if err != nil {
		panic(err)
	}

	var r io.Reader = stdout
	webm := &Webm{
		ClusterChannel: make(chan *[]byte, 1024),
		Reader: r,
	}

	
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
					/*
					err = recdvb.Start()
					if err != nil {
						fmt.Println(err)
					}
					err = ffmpeg.Start()
					if err != nil {
						fmt.Println(err)
					}
					*/
					go webm.Parse()
					
					for b := range webm.ClusterChannel {
						
						time.Sleep(50*time.Millisecond)
						cw, err := c.NextWriter(websocket.BinaryMessage)
						if err != nil {
							fmt.Println(err)
							return
						}
						cw.Write(*b)
						
						cw.Close()
					}
				}
			}
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

