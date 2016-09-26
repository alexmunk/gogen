package outputter

import "fmt"

type stdout struct{}

func (foo stdout) Send(event string) error {
	fmt.Print(event)
	return nil
}
