package t3sty

import (
  "strings"
)

// Room represents a collection of clients that all send messages to each other.
type Room struct {
  Sender        chan<- Message
  // Map of usernames to check for attendance in the Room.
  verifiedNames map[string]bool
  // Map of Client ptr's to Message channel to communicate directly to
  // each client.
  clientReceivers map[*Client]chan Message
}

// NewRoom allocates, creates, and returns a new Room.
func NewRoom() *Room {
  return &Room {
    Sender:          make(chan Message),
    verifiedNames:   make(map[string]bool),
    clientReceivers: make(map[*Client]chan Message),
  }
}

// Enter creates a new Client for a given user ready to be used.
func (rm *Room) Enter(u User) *Client {
  receiver := make(chan Message)
  client := Client {
    User:     u,
    Room:     rm,
    receiver: receiver,
  }

  rm.verifiedNames[strings.ToLower(u.Name)] = true
  rm.clientReceivers[&client] = receiver
  go client.StartListening()

  return &client
}

// Rejects the user if the username is already present.
func (rm *Room) CanEnter(u User) bool {
  _, ok := rm.verifiedNames[strings.ToLower(u.Name)]
  return !ok
}

func (rm *Room) Broadcast(msg Message) {
  for _, receiver := range rm.clientReceivers {
    receiver <- msg
  }
}
