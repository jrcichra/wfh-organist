package upnp

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"
)

type Upnp struct {
	Interface    string
	InternalPort string
	ExternalPort string
	Duration     time.Duration
	Rerun        bool
	ticker       *time.Ticker
}

// https://stackoverflow.com/a/51829730/5404881
func getInternalIP(device string) string {
	itf, _ := net.InterfaceByName(device) //here your interface
	item, _ := itf.Addrs()
	var ip net.IP
	for _, addr := range item {
		switch v := addr.(type) {
		case *net.IPNet:
			if !v.IP.IsLoopback() {
				if v.IP.To4() != nil { //Verify if IP is IPV4
					ip = v.IP
				}
			}
		}
	}
	if ip != nil {
		return ip.String()
	} else {
		return ""
	}
}

func (u *Upnp) runLoop() {
	for {
		<-u.ticker.C
		_, err := u.run()
		if err != nil {
			log.Println(err)
		}
	}
}

func (u *Upnp) Run() error {
	var err error
	u.ticker = time.NewTicker(u.Duration / 2)
	_, err = u.run()
	if u.Rerun {
		go u.runLoop() // Run every half duration period
	}
	return err
}

func (u *Upnp) run() (string, error) {
	//upnpc -a 10.0.0.126 2222 7777 udp 3 * 60
	ip := getInternalIP(u.Interface)
	if ip == "" {
		return "", fmt.Errorf("no ip found for interface %s", u.Interface)
	}
	upnpc := exec.Command("upnpc", "-a", ip, u.InternalPort, u.ExternalPort, "udp", fmt.Sprintf("%d", int(u.Duration.Seconds())))
	err := upnpc.Run()
	outp, err2 := upnpc.CombinedOutput()
	return string(outp), fmt.Errorf("%v - %v", err, err2)
}
