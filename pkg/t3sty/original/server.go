package t3sty

import (
  "io"
  "os"
  "time"
  "log"
  "strings"
  "net/http"

  "github.com/google/uuid"
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
  loginCodes map[string]User
}

func (srv *Server) generateLoginCode(u User) string {
  token := strings.ToUpper(uuid.New().String()[0:6])
  srv.loginCodes[token] = u

  go func() {
    // valid for 10 minutes.
    time.Sleep(10 * time.Minute)
    delete(srv.loginCodes, token)
  }()

  return token
}

func (srv *Server) authUser(token string) (User, bool) {
  user, prs := srv.loginCodes[strings.ToUpper(token)]
  return user, prs
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
      // has been set and connection breaks then we just log them out.
      if client != nil {
        client.Send("left chat")
        client.Leave()
      }

      break
    }

    switch msg.Type {
      case msgHello: {
        textParts := strings.Split(msg.Text, "\n")
        if len(textParts) != 2 {
          // bad hello message
          break
        }

        u := User {
          Name: textParts[0],
          Email: textParts[1],
        }

        if len(u.Name) > 120 || len(u.Email) > 120 {
          break
        }

        if srv.Room.CanEnter(u) {
          u.sendAuthEmail(srv.generateLoginCode(u))
        } else {
          conn.WriteJSON(Message {
            Type: msgMayNotEnter,
            User: u,
          })
        }
      }

      case msgAuth: {
        token := msg.Text
        u, prs := srv.authUser(token)
        if !prs {
          conn.WriteJSON(Message {
            Type: msgAuthRst,
            User: u,
          })
          break
        }

        client = srv.Room.Enter(u)
        client.OnMessage = func(msg Message) { conn.WriteJSON(msg) }

        conn.WriteJSON(Message {
          Type: msgAuthAck,
          User: u
        })

        log.Printf("@%s entered with email %s", u.Name, u.Email)

        // Sends these bot client messages ~only~ to the new client.
  			conn.WriteJSON(Message{
  				Type: msgText,
  				User: srv.BotClient.User,
  				Text: fmt.Sprintf("Hi @%s! Welcome to Plume.chat. You can read more about this project at github.com/thesephist/plume.", u.Name),
  			})
  			conn.WriteJSON(Message{
  				Type: msgText,
  				User: srv.BotClient.User,
  				Text: fmt.Sprintf("Please be kind in the chat, and remember that your email (%s) is tied to what you say here. Happy chatting!", u.Email),
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
    io.WriteSTring(w, "error reading index")
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
      loginCodes: make(map[string]User)
    }

    botUser := User{
      Name: "bot",
      Email: "bot@chat",
    }
    t3stySrv.BotClient = t3stySrv.Room.Enter(botUser)

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
