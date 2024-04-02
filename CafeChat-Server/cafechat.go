package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Channels string

const (
	HOME Channels = "home"
	QA   Channels = "qa"
	JP   Channels = "jp"
	EC   Channels = "ec"

	LogLength = 10
	SYSNAME   = "System"
)

var chatlogs = map[Channels][]Message{
	HOME: {},
	QA:   {},
	JP:   {},
	EC:   {},
}
var maidlogs = map[Channels][]MaidPrompt{
	HOME: {},
	QA:   {},
	JP:   {},
	EC:   {},
}

var addr = flag.String("addr", ":9091", "http service address")

var bans = make(map[string]bool)

func serveTestHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "../CafeChat-Production/index.html")
}

func handleFiles(w http.ResponseWriter, r *http.Request) {
	log.Println("files", r.URL, "../CafeChat-Production"+r.URL.Path)
	if strings.Contains(r.URL.Path, ".br") {
		w.Header().Set("Content-Encoding", "br")
	} else if strings.Contains(r.URL.Path, ".gz") {
		w.Header().Set("Content-Encoding", "gzip")
	}
	if strings.Contains(r.URL.Path, ".wasm") {
		w.Header().Set("Content-Type", "application/wasm")
	} else if strings.Contains(r.URL.Path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.Contains(r.URL.Path, ".gif") {
		w.Header().Set("Content-Type", "image/gif")
	} else if strings.Contains(r.URL.Path, ".png") {
		w.Header().Set("Content-Type", "image/png")
	} else if strings.Contains(r.URL.Path, ".jpg") {
		w.Header().Set("Content-Type", "image/jpeg")
	} else if strings.Contains(r.URL.Path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	http.ServeFile(w, r, "../CafeChat-Production"+r.URL.Path)
}

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()
	go serverSystem(hub)

	maidWorking.Store(HOME, false)
	maidWorking.Store(QA, false)
	maidWorking.Store(JP, false)
	maidWorking.Store(EC, false)

	http.HandleFunc("/Build/{p}", handleFiles)
	http.HandleFunc("/TemplateData/{p}", handleFiles)
	http.HandleFunc("/src/{p}", handleFiles)
	http.HandleFunc("/test", serveTestHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	http.HandleFunc("/", serveHome)
	fmt.Println("Port:", *addr)
	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serverSystem(h *Hub) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		cmd := strings.Split(text, ",")
		switch cmd[0] {
		case "c|qa":
			maidlogs[QA] = []MaidPrompt{}
			h.broadcast <- Message{send: []byte("QA maid logs cleared"), sender: SYSNAME, time: time.Now(), channel: QA}
		case "c|jp":
			maidlogs[JP] = []MaidPrompt{}
			h.broadcast <- Message{send: []byte("JP maid logs cleared"), sender: SYSNAME, time: time.Now(), channel: JP}
		case "c|ec":
			maidlogs[EC] = []MaidPrompt{}
			h.broadcast <- Message{send: []byte("EC maid logs cleared"), sender: SYSNAME, time: time.Now(), channel: EC}
		case "c":
			maidlogs = map[Channels][]MaidPrompt{
				HOME: {},
				QA:   {},
				JP:   {},
				EC:   {},
			}
			h.broadcast <- Message{send: []byte("All maid logs cleared"), sender: SYSNAME, time: time.Now(), channel: "self"}
		case "b":
			msg := strings.Trim(cmd[1], " \r\n")
			bans[msg] = true
			fmt.Println("BANNED: ", msg, " at: ", time.Now().Format("2006-01-02 15:04:05.000000000"))
		case "u":
			msg := strings.Trim(cmd[1], " \r\n")
			delete(bans, msg)
			fmt.Println("UNBAN: ", msg, " at: ", time.Now().Format("2006-01-02 15:04:05.000000000"))
		case "m":
			fmt.Println("global message")
			h.broadcast <- Message{send: []byte(strings.Trim(text[2:], " \r\n")), sender: SYSNAME, time: time.Now(), channel: "self"}
		default:
			fmt.Println("home message", cmd[0])
			if len(cmd[0]) > 1 {
				h.broadcast <- Message{send: []byte(strings.Trim(text, " \r\n")), sender: maids[HOME].name, time: time.Now(), channel: "home"}
			}
		}
	}
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast     chan Message
	maidBroadcast chan MaidMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:     make(chan Message, 10),
		maidBroadcast: make(chan MaidMessage, 10),
		register:      make(chan *Client, 10),
		unregister:    make(chan *Client, 10),
		clients:       make(map[*Client]bool, 10),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				close(client.send)
				close(client.maidSend)
				delete(h.clients, client)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					close(client.maidSend)
					delete(h.clients, client)
				}
			}
		case maidMessage := <-h.maidBroadcast:
			for client := range h.clients {
				select {
				case client.maidSend <- maidMessage:
				default:
					close(client.send)
					close(client.maidSend)
					delete(h.clients, client)
				}
			}
		}

	}
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub
	ip  string
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send     chan Message
	maidSend chan MaidMessage

	channel Channels
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		if len(message) < 2 {
			continue
		}
		if strings.Count(string(message), "\n") > 2 {
			continue
		}

		jn := strings.Split(string(message), " #")

		if jn[0] == "/join" {
			channel := strings.Trim(jn[1], " \r\n")
			switch channel {
			case "qa":
				c.channel = QA
			case "jp":
				c.channel = JP
			case "ec":
				c.channel = EC
			default:
				c.channel = HOME
			}

			message = []byte(maids[c.channel].welcomeMessage)
			log := "--Chat logs--\n"
			for _, message := range chatlogs[c.channel] {
				log += string(parseTime(message)) + "\n"
			}
			c.send <- Message{send: []byte(log), sender: SYSNAME, time: time.Now(), channel: "self"}
			// time.Sleep(3 * time.Second)
			c.maidSend <- MaidMessage{send: message, sender: maids[c.channel].name, time: time.Now(), channel: "self", emotions: []Emotion{}}
		} else if _, ok := bans[strings.Split(c.ip, ":")[0]]; !ok {
			message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
			if len(message) > 200 {
				fmt.Println(c.ip, "\t", string(message[0:200])+"...", " at: ", time.Now().Format("2006-01-02 15:04:05"))
			} else {
				fmt.Println(c.ip, "\t", string(message), " at: ", time.Now().Format("2006-01-02 15:04:05"))
			}
			c.hub.broadcast <- Message{send: message, sender: "Anonymous", time: time.Now(), channel: c.channel}
			chatlogs[c.channel] = append(chatlogs[c.channel], Message{send: message, sender: "Anonymous", time: time.Now(), channel: c.channel})
			if len(chatlogs[c.channel]) > LogLength {
				chatlogs[c.channel] = chatlogs[c.channel][1:]
			}
			if len(message) > 5 && string(message[:5]) == ".maid" && len(message) < 500 {
				beginMaidFunctions(message[5:], maids[c.channel], c.hub)
			}
		}
	}
}

// reads from associated data file to produce a message that is stored across time and can be used to change models
func beginMaidFunctions(message []byte, m Maid, hub *Hub) {

	if v, ok := maidWorking.Load(m); ok && v == true {
	} else if m.name == "verm-sama" {
		// time.Sleep(2 * time.Second)
		hub.maidBroadcast <- MaidMessage{send: []byte("I am not a maid!"), sender: m.name, time: time.Now(), channel: m.channel, emotions: []Emotion{}}
		chatlogs[m.channel] = append(chatlogs[m.channel], Message{send: []byte("I am not a maid!"), sender: m.name, time: time.Now(), channel: m.channel})
	} else {
		defer func() { maidWorking.Store(m, false) }()
		// set maidWorking[channel] = true
		maidWorking.Store(m, true)
		// read maid template from file
		// Insert maid template with message. Remove the .maid prefix
		pr, err := os.ReadFile(m.promptFile)
		if err != nil {
			panic(err)
		}
		midAction := true
		defer func() {
			midAction = false
		}()
		go func() {
			time.Sleep(1 * time.Second)
			if v, ok := maidWorking.Load(m); ok && v == true && midAction {
				switch m.name {
				case "qa-chan":
					hub.maidBroadcast <- MaidMessage{send: []byte("\"Unyah~?\""), sender: m.name, time: time.Now(), channel: m.channel, emotions: []Emotion{}}
				case "jp-chan":
					hub.maidBroadcast <- MaidMessage{send: []byte("\"Ehh...\""), sender: m.name, time: time.Now(), channel: m.channel, emotions: []Emotion{}}
				case "ec-chan":
					hub.maidBroadcast <- MaidMessage{send: []byte("\"Watashi?\""), sender: m.name, time: time.Now(), channel: m.channel, emotions: []Emotion{}}
				}

			}
		}()
		/////////////
		var prompt PromptData
		json.Unmarshal(pr, &prompt)

		ctxWindow := prompt.KissuContextWindow - 1
		basisPrompts := prompt.Messages[:len(prompt.Messages)-1]
		if len(maidlogs[m.channel]) <= ctxWindow {
			prompt.Messages = append(prompt.Messages, maidlogs[m.channel]...)
		} else {
			prompt.Messages = append(basisPrompts, maidlogs[m.channel][len(maidlogs[m.channel])-ctxWindow:]...)
		}

		cnt := string(message)
		cnt = strings.ReplaceAll(cnt, "\"", "'")
		re := regexp.MustCompile(`[^一-龠ぁ-ゔァ-ヴーa-zA-Z0-9ａ-ｚＡ-Ｚ０-９々〆〤ヶ!&+~.,/?:' ]`)
		re2 := regexp.MustCompile(`  `)
		cnt = re.ReplaceAllString(cnt, "")
		cnt = re2.ReplaceAllString(cnt, " ")
		cnt = strings.Trim(cnt, " \r\n")
		prompt.Messages = append(prompt.Messages, MaidPrompt{
			Role:    "user",
			Content: "\"" + cnt + "\"",
		})

		maidlogs[m.channel] = append(maidlogs[m.channel], prompt.Messages[len(prompt.Messages)-1])
		prompt.Messages[len(prompt.Messages)-1].Content += prompt.PromptEnder
		///////////

		// Get requested model from file
		// Make request to ST on 127.0.0.1:8000 with maid data
		// Anthropic token is handled in ST
		// convert into emotion array
		client := &http.Client{}
		b, err := json.Marshal(prompt)
		if err != nil {
			panic(err)
		}
		greq, _ := http.NewRequest("POST", "http://127.0.0.1:8000/api/backends/chat-completions/generate", bytes.NewBuffer(b))
		for k, v := range genericHeaders {
			greq.Header.Add(k, v)
		}
		res, err := client.Do(greq)

		midAction = false

		if err != nil {
			fmt.Println("ERROR ON REQUEST: ", err)
			return
		}
		defer res.Body.Close()
		var maidResponse GenerationResponse
		if res.Header.Get("Content-Encoding") == "gzip" {
			var gzipReader *gzip.Reader
			gzipReader, err = gzip.NewReader(res.Body)
			if err != nil {
				fmt.Println("ERROR ON GZIP: ", err)
				return
			}
			err = json.NewDecoder(gzipReader).Decode(&maidResponse)
		} else {
			err = json.NewDecoder(res.Body).Decode(&maidResponse)
		}
		if err != nil {
			fmt.Println("ERROR ON DECODE: ", err)
			return
		}
		fmt.Println("MAID RESPONSE:", maidResponse)
		if len(maidResponse.Choices) == 0 {
			fmt.Println("NO CHOICES")
			hub.maidBroadcast <- MaidMessage{send: []byte("I'm sorry, but my brain went 404... Can you repeat yourself?"), sender: m.name, time: time.Now(), channel: m.channel, emotions: []Emotion{}}
			return
		}
		mCnt := strings.Trim(string(maidResponse.Choices[0].Message.Content), " \r\n")
		var maidMessage MaidMessage = MaidMessage{
			send:     []byte(mCnt),
			sender:   m.name,
			time:     time.Now(),
			channel:  m.channel,
			emotions: []Emotion{},
		}
		maidlogs[m.channel] = append(maidlogs[m.channel], MaidPrompt{
			Role:    "assistant",
			Content: mCnt,
		})

		client = &http.Client{}
		emotionPrompt := EmotionData{Text: string(maidResponse.Choices[0].Message.Content)}
		b, err = json.Marshal(emotionPrompt)
		if err != nil {
			fmt.Println("EMOTION MARSHAL", err)
			return
		}
		ereq, _ := http.NewRequest("POST", "http://127.0.0.1:8000/api/extra/classify", bytes.NewBuffer(b))
		for k, v := range genericHeaders {
			ereq.Header.Add(k, v)
		}
		res, err = client.Do(ereq)
		if err != nil {
			fmt.Println("ERROR ON REQUEST: ", err)
			return
		}
		defer res.Body.Close()
		var emotionsResponse EmotionsResponse
		err = json.NewDecoder(res.Body).Decode(&emotionsResponse)
		if err != nil {
			fmt.Println("ERROR ON DECODE: ", err)
			return
		}
		maidMessage.emotions = emotionsResponse.Classification
		// broadcast response to chanel
		chatlogs[m.channel] = append(chatlogs[m.channel], Message{maidMessage.send, m.name, time.Now(), m.channel})
		hub.maidBroadcast <- maidMessage
	}

}

func parseTime(message Message) []byte {
	return []byte(fmt.Sprintf("%s[%s]: %s", message.sender, message.time.Format(time.TimeOnly), message.send))
}
func parseMessage(message Message) []byte {
	return []byte(fmt.Sprintf("%s: %s", message.sender, message.send))
}

func parseMaidMessage(message MaidMessage) []byte {
	e := []string{}
	for _, em := range message.emotions {
		e = append(e, fmt.Sprintf("%s&%f", em.Label, em.Score))
	}
	eStr := strings.Join(e, ",")
	return []byte(fmt.Sprintf("%s||%s: %s", message.sender, eStr, message.send))
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	observingMM := true
	observingCM := true
	for observingMM || observingCM {
		select {
		case maidMessage, ok := <-c.maidSend:
			channel := maidMessage.channel
			observingMM = true
			if !ok {
				observingMM = false
				continue
			} else if channel != c.channel && channel != "self" {
				continue
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(parseMaidMessage(maidMessage))

			// Add queued chat messages to the current websocket message.
			n := len(c.maidSend)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(parseMaidMessage(<-c.maidSend))
			}

			if err := w.Close(); err != nil {
				return
			}

		case message, ok := <-c.send:
			channel := message.channel
			observingCM = true
			if !ok {
				// The hub closed the channel.
				observingCM = false
				continue
			} else if channel != c.channel && channel != "self" {
				continue
			}
			writeFn := parseMessage
			if message.sender == SYSNAME {
				writeFn = parseTime
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(writeFn(message))

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(writeFn(<-c.send))
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	// read get request from r
	channel := r.URL.Query().Get("channel")

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}
	ip := r.Header.Get("X-Real-IP")
	client := &Client{hub: hub, conn: conn, send: make(chan Message, 10), maidSend: make(chan MaidMessage, 10), channel: HOME, ip: ip}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()

	switch channel {
	case "qa":
		client.channel = QA
	case "jp":
		client.channel = JP
	case "ec":
		client.channel = EC
	default:
		client.channel = HOME
	}

	client.hub.register <- client
	fmt.Println("connected: ", ip, " at: ", time.Now().Format("2006-01-02 15:04:05.000000000"))

	client.send <- Message{send: []byte("--Chat logs--"), sender: SYSNAME, time: time.Now(), channel: "self"}
	for _, message := range chatlogs[client.channel] {
		client.send <- Message{send: message.send, sender: message.sender, time: message.time, channel: "self"}
	}
	client.maidSend <- MaidMessage{send: []byte(maids[client.channel].welcomeMessage), sender: maids[client.channel].name, time: time.Now(), channel: "self", emotions: []Emotion{}}
}

////////////////////////
////////////////////////
////////////////////////
////////////////////////
////////////////////////
////////////////////////

var maids = map[Channels]Maid{
	HOME: {
		name:           "verm-sama",
		welcomeMessage: "Welcome to Kissu-3.0. The new chat engine with <u>maids</u>! Address the maids with <b>'.maid'</b> and they will respond to you!",
		channel:        HOME,
		promptFile:     "",
	},
	QA: {
		name:           "qa-chan",
		welcomeMessage: "\"Unyah~!? You started me, nyaa!\" qa-chan jumps as she is distracted from writing a lengthy imageboard post about her hobbies. \"Can I help you, Anonyamous?\"",
		channel:        QA,
		promptFile:     "qa.json",
	},
	JP: {
		name:           "jp-chan",
		welcomeMessage: "\"Ehhhh? What do you want, ojisan? Can't you see I'm playin' some Touhou?\" the bratty girl speaks dismissively as Reimu explodes on screen and jp-chan runs out of lives. \"HMPH! Look at what you made me do! ( · `ω´ · )\"",
		channel:        JP,
		promptFile:     "jp.json",
	},
	EC: {
		name:           "ec-chan",
		welcomeMessage: "\"Itterasshai Anonymous~ Take a seat and order some shochu!\" She leans forward slightly \"Or maybe you'd like something softer?\"",
		channel:        EC,
		promptFile:     "ec.json",
	},
}

var maidWorking = sync.Map{}
var genericHeaders = map[string]string{
	"Host":            "127.0.0.1:8000",
	"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0",
	"Accept":          "*/*",
	"Accept-Language": "en-CA,en-US;q=0.7,en;q=0.3",
	"Accept-Encoding": "gzip, deflate, br",
	"Referer":         "http://127.0.0.1:8000/",
	"Content-Type":    "application/json",
	"X-CSRF-Token":    "disabled",
	"Origin":          "http://127.0.0.1:8000",
	"Connection":      "keep-alive",
	"Sec-Fetch-Dest":  "empty",
	"Sec-Fetch-Mode":  "cors",
	"Sec-Fetch-Site":  "same-origin",
	"Pragma":          "no-cache",
	"Cache-Control":   "no-cache",
}

type Message struct {
	send    []byte
	sender  string
	time    time.Time
	channel Channels
}

type MaidMessage struct {
	send     []byte
	sender   string
	time     time.Time
	channel  Channels
	emotions []Emotion
}

type Emotion struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

type Maid struct {
	name           string
	welcomeMessage string
	channel        Channels
	promptFile     string
}

type MaidPrompt struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type PromptData struct {
	Messages    []MaidPrompt `json:"messages"`
	Model       string       `json:"model"`
	Temperature float64      `json:"temperature"`
	Frequency   float64      `json:"frequency_penalty"`
	Presence    float64      `json:"presence_penalty"`
	TopP        float64      `json:"top_p"`
	MaxTokens   int          `json:"max_tokens"`
	Stream      bool         `json:"stream"`
	LogitBias   struct {
	} `json:"logit_bias"`
	ChatCompletionSource  string `json:"chat_completion_source"`
	UserName              string `json:"user_name"`
	CharName              string `json:"char_name"`
	TopK                  int    `json:"top_k"`
	ClaudeUseSysprompt    bool   `json:"claude_use_sysprompt"`
	Stop                  []string
	HumanSyspromptMessage string `json:"human_sysprompt_message"`
	AssistantPrefill      string `json:"assistant_prefill"`

	KissuContextWindow int    `json:"kissu_ctx_window"`
	PromptEnder        string `json:"prompt_ender"`
}

type GenerationResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type EmotionData struct {
	Text string `json:"text"`
}

type EmotionsResponse struct {
	Classification []Emotion `json:"classification"`
}
