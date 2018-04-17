package main

import (
	"./fanotify"
	"./utils"
	"./clamav"
	"fmt"
  "strconv"
  "runtime"
  gocache "github.com/pmylund/go-cache"
)

const (
	AT_FDCWD = -100
)

var (
  log utils.Logger

  cfg = utils.NewConfig()
  debug, _ = strconv.ParseBool(cfg.Options["log"]["debug_enabled"])
  num_cpus, _ = strconv.Atoi(cfg.Options["scan"]["num_cpus"])
  num_routines, _ = strconv.Atoi(cfg.Options["scan"]["num_routines"])

  // AV specific variables
  av_mount_point = cfg.Options["scan"]["mount_point"]
  av_action = cfg.Options["scan"]["av_action_when_malware_found"]
  av_max_file_size, _  = strconv.ParseInt(cfg.Options["scan"]["av_max_file_size"], 10, 64)
  av_quarantine_dir = cfg.Options["scan"]["av_move_to_base_folder"]

  // AMQP
  amqp_enabled, _ = strconv.ParseBool(cfg.Options["amqp"]["amqp_enabled"])
)


func clamavWorker(in chan *fanotify.EventMetadata, cache *gocache.Cache, number int){

  log.Debug(fmt.Sprintf("[%d] initializing ClamAV database...", number), debug)

  engine := clamav.New()
	sigs, err := engine.Load(clamav.DBDir(), clamav.DbStdopt)
  utils.CheckPanic(err, fmt.Sprintf("can not initialize ClamAV engine: %v", err))

  engine.Compile()
	log.Debug(fmt.Sprintf("loaded %d signatures", sigs), debug)


  for ev := range in {
    log.Debug(fmt.Sprintf("[%d] File: %s, Size %d", number, ev.FileName, ev.Size), debug)

    _, found := cache.Get(fmt.Sprintf("%d", ev.InodeNumber))

    if ((!found || ev.Mask == fanotify.FAN_CLOSE_WRITE) && !ev.IsDir && ev.IsRegular && ev.Size > 0 && ev.Size < av_max_file_size){
      cache.Set(fmt.Sprintf("%d", ev.InodeNumber), 1, -1)

      log.Debug(fmt.Sprintf("[%d] Scanning %s...", number, ev.FileName), debug)
  		virus, _, err := engine.ScanFile(ev.FileName, clamav.ScanStdopt)
  		if virus != "" {
        log.Log(fmt.Sprintf("[%d] malware found in %s: %s", number, ev.FileName, virus))
        if av_action == "MOVE"{
          err := utils.MoveFile(ev.FileName, av_quarantine_dir, log)
          if err != nil{
            cache.Delete(fmt.Sprintf("%d", ev.InodeNumber))
          }
        } else if av_action == "REMOVE" {
          err := utils.RemoveFile(ev.FileName, log)
          if err != nil{
            cache.Delete(fmt.Sprintf("%d", ev.InodeNumber))
          }
        }

  		} else if err != nil {
        log.Debug(fmt.Sprintf("[%d] Error scanning file %s: %v", number, ev.FileName, err), debug)
        cache.Delete(fmt.Sprintf("%d", ev.InodeNumber))
        continue
      }
    }
  }

}

func main() {
  log.SetupLogger(amqp_enabled)

  runtime.GOMAXPROCS(num_cpus)

	fan, err := fanotify.Initialize(fanotify.FAN_CLASS_NOTIF, fanotify.FAN_CLOEXEC)
	utils.CheckPanic(err, "Unable to listen fanotify")
	fan.Mark(fanotify.FAN_MARK_ADD|fanotify.FAN_MARK_MOUNT, fanotify.FAN_CLOSE, AT_FDCWD, av_mount_point)

  cache := gocache.New(0, 0)
  channel := make(chan *fanotify.EventMetadata)

  for i := 0; i < num_routines; i++ {
    go clamavWorker(channel, cache, i)
  }

	for {
		ev, err := fan.GetEvent()
    if err == nil {
      channel <- ev
    } else {
      log.Debug(fmt.Sprintf("Fanotify error: %v", err), debug)
      continue
    }
	}

}
