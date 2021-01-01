package t3sty

type Error struct {
  reason string
}

func (err Error) Error() string {
  return err.reason
}
