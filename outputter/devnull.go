package outputter

type devnull struct{}

func (foo devnull) Send(event string) error {
	return nil
}
