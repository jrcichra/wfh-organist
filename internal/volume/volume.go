package volume

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

var curVol = -1

func SetVolume(volume int) error {
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}

	if curVol != volume {
		curVol = volume
		// get current script's directory
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		path := filepath.Dir(exe)
		log.Printf("Setting volume to: %d%%\n", volume)
		return exec.Command(path+"/volume.sh", strconv.Itoa(volume)).Run()
	}

	return nil
}
