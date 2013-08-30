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
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"strings"
)

type Config struct {
	Options map[string]map[string]string
}

func NewConfig() Config {
	cfg := Config{}

	config_file := os.Getenv("HPR_CONF")
	if strings.EqualFold(config_file, "") {
		config_file = "/etc/hpr/hpr.yml"
	}
	f, err := ioutil.ReadFile(config_file)
	CheckPanic(err, "Unable to open YAML file")

	err = goyaml.Unmarshal(f, &cfg.Options)
	CheckPanic(err, "Unable to parse YAML file")
	return cfg
}
