package broker

import (
	"os"
)

// fs is dummy broker which saves every request into the same file.

const openFlags = os.O_CREATE | os.O_APPEND | os.O_WRONLY

// this will save payloads to the file with name = {name}
func GetAsyncFsSendFunc(name string) AsyncSendFunc {
	return func(p []Payload) ([]Payload, error) {
		fd, err := os.OpenFile(name, openFlags, 0600)
		if err != nil {
			return p, err
		}
		defer fd.Close()
		for i := range p {
			if _, err := fd.Write(p[i].Bytes); err != nil {
				return p[i:], err
			}
		}
		return nil, nil
	}
}

// this will save payloads to the file with name = {name}
func GetFsSendFunc(name string) SendFunc {
	return func(p *Payload) error {
		fd, err := os.OpenFile(name, openFlags, 0600)
		if err != nil {
			return err
		}
		defer fd.Close()
		_, err = fd.Write(p.Bytes)
		return err
	}
}
