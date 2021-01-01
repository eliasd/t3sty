package t3sty

const (
  // Sent by client to first connect and request authentication.
  msgHello = iota

  // Sent by client and server for regular text exchanges.
  msgText

  // Sent by client to attempt to authenticate with a token (with the
  // server).
  msgAuth
  // Sent by server to approve an authentication attempt.
  msgAuthAck
  // Sent by server to reject an authentication attempt.
  msgAuthRst
  // Sent by server to reject a client's entry attempt into the room
  // (usually means the username is taken).
  msgMayNotEnter
)

// Struct for communication between a client and the server.
type Message struct {
  Type int
  User User
  Text string
}
