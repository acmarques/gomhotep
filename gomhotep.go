package main

import (
	"./fanotify"
	"./utils"
	"./clamav"  
	"fmt"
  "os"
  "strconv"
  "runtime"
  gocache "github.com/pmylund/go-cache"
)

const (
	AT_FDCWD = -100
)



func clamavWorker(in chan *fanotify.EventMetadata, cache *gocache.Cache, cfg utils.Config, number int){
  debug, err := strconv.ParseBool(cfg.Options["log"]["debug_enabled"])
  utils.CheckPanic(err, "Unable to get debug configuration")
  
  utils.Debug(fmt.Sprintf("[%d] initializing ClamAV database...", number), debug)
  engine := clamav.New()
	sigs, err := engine.Load(clamav.DBDir(), clamav.DbStdopt)
  utils.CheckPanic(err, fmt.Sprintf("can not initialize ClamAV engine: %v", err))
	engine.Compile()  
	utils.Debug(fmt.Sprintf("loaded %d signatures", sigs), debug)
  
  
  for ev := range in {  
    fi, err := os.Stat(ev.FileName) 
    utils.CheckPanic(err, fmt.Sprintf("Error getting file information %s: %v", ev.FileName, err)) 
    
    _, found := cache.Get(fmt.Sprintf("%d", ev.InodeNumber))
  
    if (!found && !fi.IsDir()){
      utils.Debug(fmt.Sprintf("[%d] Scanning %s...", number, ev.FileName), debug)
  		virus, _, err := engine.ScanFile(ev.FileName, clamav.ScanStdopt)
  		if virus != "" {
  			utils.Log(fmt.Sprintf("[%d] virus found in %s: %s", number, ev.FileName, virus))
  		} else if err != nil {
        utils.CheckPanic(err, fmt.Sprintf("error scanning %s: %v", ev.FileName, err))
      }
      cache.Set(fmt.Sprintf("%d", ev.InodeNumber), 1, -1)
    }
  }
}

func main() {
  cfg := utils.NewConfig()
  
  num_cpus, err := strconv.Atoi(cfg.Options["scan"]["num_cpus"])
  utils.CheckPanic(err, "Unable to parse configuration: number of cpus to use")
  runtime.GOMAXPROCS(num_cpus)
  
	fan, err := fanotify.Initialize(fanotify.FAN_CLASS_NOTIF, fanotify.FAN_CLOEXEC)
	utils.CheckPanic(err, "Unable to listen fanotify")
	fan.Mark(fanotify.FAN_MARK_ADD|fanotify.FAN_MARK_MOUNT, fanotify.FAN_CLOSE, AT_FDCWD, cfg.Options["scan"]["mount_point"])

  cache := gocache.New(0, 0)
  channel := make(chan *fanotify.EventMetadata)
  
  num_routines, err := strconv.Atoi(cfg.Options["scan"]["num_routines"])
  utils.CheckPanic(err, "Unable to parse configuration: number of scanning routines")
  for i := 0; i < num_routines; i++ {
    go clamavWorker(channel, cache, cfg, i)
  }

	for {
		ev, _ := fan.GetEvent()
    channel <- ev
	}
}
