package rx

import (
	"log"
	"os/exec"
)

// manage rx process

type Rx struct {
	Port    string
	command *exec.Cmd
}

func (rx *Rx) Kill() error {
	if rx.command != nil {
		err := rx.command.Process.Kill()
		if err != nil {
			log.Println(err)
			return err
		}
		rx.command = nil
	}
	return nil
}

func (rx *Rx) Run() error {

	if rx.command != nil {
		err := rx.Kill()
		if err != nil {
			return err
		}
	}

	if rx.Port == "" {
		// default to 1350
		rx.Port = "1350"
	}

	// run rx command
	rx.command = exec.Command("rx", "-p", rx.Port)
	err := rx.command.Run() // blocks until it is complete
	if err != nil {
		return err
	}
	return nil
}
