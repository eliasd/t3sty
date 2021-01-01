package t3sty

import (
  "strings"
)

// Room represents a collection of clients that all send messages to each
// other.
type Room struct {
  Sender          chan<- Message
  // Map of usernames to emails.
  verifiedNames   map[string]string
  // Map of Client ptr's to Message channel.
  clientReceivers map[*Client]chan Message
}

// NewRoom allocates, creates, and returns a new Room.
func NewRoom() *Room {
  return &Room {
    Sender:          make(chan Message),
    verifiedNames:   make(map[string]string),
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

  rm.verifiedNames[strings.ToLower(u.Name)] = u.Email
  rm.clientReceivers[&client] = receiver
  go client.StartListening()

  return &client
}

func (rm *Room) CanEnter(u User) bool {
	existingEmail, prs := rm.verifiedNames[strings.ToLower(u.Name)]
	if prs {
		return u.Email == existingEmail
	}

	return true
}

func (rm *Room) Broadcast(msg Message) {
  for _, receiver := range rm.clientReceivers {
    receiver <- msg
  }
}
