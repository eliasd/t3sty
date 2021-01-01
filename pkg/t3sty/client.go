package t3sty

type Client struct {
  User      User
  Room      *Room
  OnMessage func(Message)
  receiver  chan Message
}

// Send sends a new message to the room (and all the clients within the room)
// to which the client cl belongs to, under the client user's name.
func (cl *Client) Send(text string) error {
  if cl.Room == nil {
    return Error{"client is not in a room yet"}
  }

  cl.Room.Broadcast(Message {
    Type: msgText,
    User: cl.User,
    Text: text,
  })

  return nil
}

// Leave lets the client leave the room and cleans up.
func (cl *Client) Leave() error {
  if cl.Room == nil {
    return Error{"client is not in a room yet"}
  }

  delete(cl.Room.verifiedNames, cl.User.Name)
  delete(cl.Room.clientReceivers, cl)
  close(cl.receiver)
  cl.Room = nil

  return nil
}

// StartListening listens for new messages for the client (from the client
// receiver to the Room) and sends them to the client with cl.OnMessage(msg).
// Stops listening once the client's receiver is closed.
func (cl *Client) StartListening() {
  for {
    msg, open := <-cl.receiver
    if !open { return }                   // Quit listening for client if the channel closes.
    if cl.OnMessage == nil { continue }   // Skips if OnMessage is not yet set.

    cl.OnMessage(msg)
  }
}
