package t3sty

import (
  "io"
  "os"
  "time"
  "log"
  "strings"
  "net/http"

  "github.com/gorilla/mux"
  "github.com/gorilla/websocket"
  "golang.org/x/time/rate"
)

const maxTextLen = 65536

var upgrader = websocket.Upgrader {
      ReadBufferSize: 1024,
      WriteBufferSize: 1024,
      CheckOrigin: func (r *http.Request) bool {
        origin := r.Header.Get("Origin")
        return origin == "https://t3sty.dev" || origin == "https://www.t3sty.dev"
      },
}

type Server struct {
  Room       *Room
  BotClient  *Client
}

// Server executes connect for each new connection to the server, creating a
// new Websocket that corresponds to the client's newly created Websocket.
func (srv *Server) connect(w http.ResponseWriter, r *http.Request) {
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Println(err)
    return
  }
  defer conn.Close()

  var client *Client

  // concurrent func keeps the websocket connection to client alive.
  go func() {
    for {
      // 50 seconds, as the HTTP Read / Write timeout is 60 seconds for
      // this server.
      time.Sleep(50 * time.Second)  //

      if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
        if client != nil {
          client.Leave()
        }
        return
      }
    }
  }()

  // Continuously runs server operations for the new connection until the
  // the connection breaks (which occurs once the client closes the page).
  messageLimiter := rate.NewLimiter(10, 1)
  for {
    var msg Message

    err := conn.ReadJSON(&msg)
    if err != nil {
      log.Printf("error: %v", err)

      // Websocket connection can error for multiple reasons but if the client
      // has been set and connection breaks then we just log them out / clean up.
      if client != nil {
        client.Send("left chat")
        client.Leave()
      }

      break
    }

    switch msg.Type {
      case msgHello: {
        u := User {
          Name: msg.Text,
        }

        // Server sends msgMayNotEnter:
        // Server rejects client if user name is too long or user can't enter Room
        // because the user name is taken.
        if len(u.Name) > 120 || !srv.Room.CanEnter(u) {
          conn.WriteJSON(Message {
            Type: msgMayNotEnter,
            User: u,
          })
        }

        client = serv.Room.Enter(u)
        client.OnMessage = func(msg Message) { conn.WriteJSON(msg) }

        // Server sends msgAuthAck:
        // Server notifies the client that they've been entered into the Room.
        conn.WriteJSON(Message {
          Type: msgAuthAck,
          User: u,
        })

        log.Printf("@%s entered room.", u.Name)

        // Sends these bot client messages ~only~ to the new client.
        conn.WriteJSON(Message {
  				Type: msgText,
  				User: srv.BotClient.User,
  				Text: fmt.Sprintf("Hi @%s! Welcome to t3sty.chat. It's an educational partial re-implementation of github.com/thesephist/plume.", u.Name),
        })

        client.Send("entered chat")
      }

      case msgText: {
        if client == nil { break }
        if !messageLimiter.Allow() { break }
        if len(msg.Text) > maxTextLen { msg.Text = msg.Text[0:maxTextLen] }
        client.Send(msg.Text)
      }

      default:
        log.Printf("unknown message type: %v", msg)
    }
  }
}

func handleHome(w http.ResponseWriter, r *http.Request) {
  // Needs to be run in this relative path.
  indexFile, err := os.Open("./static/index.html")
  defer indexFile.Close()

  if err != nil {
    io.WriteString(w, "error reading index")
    return
  }

  io.Copy(w, indexFile)
}

func StartServer() {
    // Initiate http request router / dispatcher.
    r := mux.NewRouter()

    // Set up http server.
    srv := &http.Server {
      Handler:      r,
      Addr:         "127.0.0.1:9990",
      WriteTimeout: 60 * time.Second,
      ReadTimeout:  60 * time.Second,
    }

    // Set up chat server config info.
    t3stySrv := Server {
      Room: NewRoom(),
    }
    t3stySrv.BotClient = t3stySrv.Room.Enter(User { Name: "bot", })

    // Configure http request router.
    r.HandleFunc("/", handleHome)
    r.PathPrefix("/static/").Handler(
      http.StripPrefix("/static/", http.FileServer(http.Dir("./static")))
    )
    r.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
      t3stySrv.connect(w, r)
    })

    // Start http server
    log.Printf("Server listening on %s\n", srv.Addr)
    log.Fatal(srv.ListenAndServe())
}
