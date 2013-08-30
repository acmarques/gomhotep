package main

import (
	"./fanotify"
	gomohtep_utils "./utils"
	"fmt"
)

type Directories struct {
	Directory []string
}

const (
	AT_FDCWD = -100
)

func main() {
	fan, err := fanotify.Initialize(fanotify.FAN_CLASS_NOTIF, fanotify.FAN_CLOEXEC)
	gomohtep_utils.CheckPanic(err, "Unable to listen fanotify")

	fan.Mark(fanotify.FAN_MARK_ADD|fanotify.FAN_MARK_MOUNT, fanotify.FAN_CLOSE_WRITE, AT_FDCWD, "/home/ncode")

	for {
		ev, _ := fan.GetEvent()
		fmt.Println(ev)
	}
}
