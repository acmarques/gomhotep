/*
   Copyright 2013 Juliano Martinez <juliano@martinez.io>

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

   @author: Juliano Martinez
*/

package utils

import (
	"fmt"
	"log"
	"log/syslog"
)

var (
	_log, s_err = syslog.New(syslog.LOG_ERR, "hpr")
)

func Check(err error, message string) {
	check(err, message, false)
}

func CheckPanic(err error, message string) {
	check(err, message, true)
}

func Log(message string) {
	CheckPanic(s_err, "Unable to write syslog message")
	_log.Info(message)
	defer _log.Close()
}

func check(err error, message string, _panic bool) {
	if err != nil {
		msg := fmt.Sprintf("%s: %s", message, err)
		if s_err != nil {
			log.Fatalln("Unable to write syslog message")
		}
		_log.Warning(msg)
		defer _log.Close()
		log.Fatalln(msg)
		if _panic {
			panic(msg)
		}
	}
}
