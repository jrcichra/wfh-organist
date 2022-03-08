package tx

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/jrcichra/wfh-organist/pkg/upnp"
)

// manage tx process

type Tx struct {
	Address     string
	Interface   string
	Port        string
	DisableUpnp bool
	command     *exec.Cmd
}

func (tx *Tx) Kill() error {
	if tx.command != nil {
		err := tx.command.Process.Kill()
		if err != nil {
			log.Println(err)
			return err
		}
		tx.command = nil
	}
	return nil
}

func (tx *Tx) Run() error {

	if tx.command != nil {
		err := tx.Kill()
		if err != nil {
			return err
		}
	}
	if tx.Address == "" {
		return fmt.Errorf("no address specified")
	}
	if tx.Port == "" {
		// default to 1350
		tx.Port = "1350"
	}
	if !tx.DisableUpnp {
		u := upnp.Upnp{
			Interface:    tx.Interface,
			InternalPort: tx.Port,
			ExternalPort: tx.Port,
			Duration:     time.Hour * 5,
			Rerun:        true,
		}
		err := u.Run()
		if err != nil {
			return err
		}
	}
	// run tx command
	tx.command = exec.Command("tx", "-h", tx.Address, "-p", tx.Port)
	err := tx.command.Run() // blocks until it is complete
	if err != nil {
		return err
	}
	return nil
}
