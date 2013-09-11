package main

import (
	"./fanotify"
  "./clamav"
	"./utils"
	"fmt"
  "github.com/pmylund/go-cache"
)

type Directories struct {
	Directory []string
}

const (
	AT_FDCWD = -100
)


func initClamAV() *clamav.Engine {
	engine := clamav.New()
	sigs, err := engine.Load(clamav.DBDir(), clamav.DbStdopt)
	if err != nil {
		fmt.Println("can not initialize ClamAV engine: %v", err)
	}

	fmt.Println(fmt.Sprintf("loaded %d signatures", sigs))

	engine.Compile()

	return engine
}


func main() {
  var engine *clamav.Engine

  fmt.Println("initializing ClamAV database...")
  engine = initClamAV()

	fan, err := fanotify.Initialize(fanotify.FAN_CLASS_NOTIF, fanotify.FAN_CLOEXEC)
	utils.CheckPanic(err, "Unable to listen fanotify")

	fan.Mark(fanotify.FAN_MARK_ADD|fanotify.FAN_MARK_MOUNT, fanotify.FAN_LW_EVENTS, AT_FDCWD, "/home/ncode")

  c := cache.New(0, 0)

	for {
		ev, _ := fan.GetEvent()
    fmt.Println(ev)
		  
    _, found := c.Get(fmt.Sprintf("%d", ev.InodeNumber))
    if (!found){
      fmt.Println(fmt.Sprintf("Scanning %s...", ev.FileName))
  		virus, _, err := engine.ScanFile(ev.FileName, clamav.ScanStdopt)
  		if virus != "" {
  			fmt.Println(fmt.Sprintf("virus found in %s: %s", ev.FileName, virus))
  		} else if err != nil {
  			fmt.Println(fmt.Sprintf("error scanning %s: %v", ev.FileName, err))
  		}    
      c.Set(fmt.Sprintf("%d", ev.InodeNumber), 1, -1)
    }
	}
}
