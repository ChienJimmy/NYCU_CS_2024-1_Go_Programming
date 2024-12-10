package main

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/reactivex/rxgo/v2"
)

type client chan<- string // an outgoing message channel

var (
	entering      = make(chan client)
	leaving       = make(chan client)
	messages      = make(chan rxgo.Item) // all incoming client messages
	ObservableMsg = rxgo.FromChannel(messages)
)

func broadcaster() {
	clients := make(map[client]bool) // all connected clients
	MessageBroadcast := ObservableMsg.Observe()
	for {
		select {
		case msg := <-MessageBroadcast:
			// Broadcast incoming message to all
			// clients' outgoing message channels.
			for cli := range clients {
				cli <- msg.V.(string)
			}

		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func clientWriter(conn *websocket.Conn, ch <-chan string) {
	for msg := range ch {
		conn.WriteMessage(1, []byte(msg))
	}
}

func wshandle(w http.ResponseWriter, r *http.Request) {
	upgrader := &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	ch := make(chan string) // outgoing client messages
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	ch <- "你是 " + who + "\n"
	messages <- rxgo.Of(who + " 來到了現場" + "\n")
	entering <- ch

	defer func() {
		log.Println("disconnect !!")
		leaving <- ch
		messages <- rxgo.Of(who + " 離開了" + "\n")
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		messages <- rxgo.Of(who + " 表示: " + string(msg))
	}
}

func InitObservable() {
	// TODO: Please create an Observable to handle the messages
	/*
		ObservableMsg = ObservableMsg.Filter(...) ... {
		}).Map(...) {
			...
		})
	*/
	// Load swear words and sensitive names
	swearWords := loadWords("swear_word.txt")
	sensitiveNames := loadWords("sensitive_name.txt")

	ObservableMsg = ObservableMsg.
		Filter(func(item interface{}) bool {
			msg := item.(string)
			for _, word := range swearWords {
				if strings.Contains(msg, word) {
					return false // Filter out messages containing swear words
				}
			}
			return true
		}).
		Map(func(_ context.Context, item interface{}) (interface{}, error) {
			msg := item.(string)
			for _, name := range sensitiveNames {
				if strings.Contains(msg, name) {
					// Replace the second character of sensitive names with '*'
					runes := []rune(name)
					if len(runes) > 1 {
						maskedName := string(runes[0]) + "*" + string(runes[2:])
						msg = strings.ReplaceAll(msg, name, maskedName)
					}
				}
			}
			return msg, nil
		})
}

// Helper function to load words from a file
func loadWords(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Failed to read file %s: %v", filename, err)
	}
	return words
}

func main() {
	InitObservable()
	go broadcaster()
	http.HandleFunc("/wschatroom", wshandle)

	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("server start at :8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
