/* Copyright 2013 Juliano Martinez
   All Rights Reserved.

     Licensed under the Apache License, Version 2.0 (the "License");
     you may not use this file except in compliance with the License.
     You may obtain a copy of the License at

         http://www.apache.org/licenses/LICENSE-2.0

     Unless required by applicable law or agreed to in writing, software
     distributed under the License is distributed on an "AS IS" BASIS,
     WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
     See the License for the specific language governing permissions and
     limitations under the License.

   @author: Juliano Martinez */

package utils

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"sync"
	"fmt"
	"os"
	"time"
  "strconv"
)

var (
  cfg = NewConfig()
  amqp_uri = cfg.Options["amqp"]["amqp_url"]
  amqp_queue = cfg.Options["amqp"]["amqp_queue"]
  amqp_message_ttl = cfg.Options["amqp"]["message_ttl"]
)


type AMQPConnection struct {
	conn       *amqp.Connection
	pub        *amqp.Channel
	Queue      string
	Expiration string
	Uri        string
	mu         sync.RWMutex
}

type Graylog2Parsed struct {
	Host         string `json:"host"`
	Timestamp    int64  `json:"timestamp"`
	Facility     string `json:"facility"`
	Level        int    `json:"level"`
	Version      string `json:"version"`
	ShortMessage string `json:"short_message"`
}

var Severity = []string{"emerg", "alert", "crit", "err", "warn", "notice", "info", "debug"}
var Facility = []string{"kern", "user", "mail", "daemon", "auth", "syslog", "lpr",
	"news", "uucp", "cron", "authpriv", "ftp", "ntp", "audit",
	"alert", "at", "local0", "local1", "local2", "local3",
	"local4", "local5", "local6", "local7"}


func Graylog2ParseLog(line string) Graylog2Parsed {
  	parsed := Graylog2Parsed{}
  	now := time.Now()
  	parsed.Timestamp = now.Unix()
  	parsed.Version = "1.0"
  	hostname, err := os.Hostname()
  	CheckPanic(err, fmt.Sprintf("Unable to get my hostname"))
  	parsed.Host = hostname
  	
		parsed.Facility = "Syslog"
		parsed.Level = 6
		parsed.ShortMessage = line

    return parsed
  }

func setup(uri, queue string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		Check(err, "Unable to connect to broker")
		return nil, nil, err
	}

	pub, err := conn.Channel()
	if err != nil {
		Check(err, "Unable to acquire channel")
		return nil, nil, err
	}

	err = pub.ExchangeDeclare(queue, "direct", true, false, false, false, nil)
	if err != nil {
		Check(err, "Unable to declare exchange")
		return nil, nil, err
	}

  message_ttl, _ := strconv.Atoi(amqp_message_ttl)
  args := amqp.Table{"x-message-ttl": message_ttl}
	_, err = pub.QueueDeclare(queue, true, false, false, false, args)
	if err != nil {
		Check(err, "Unable to declare queue")
		return nil, nil, err
	}

	err = pub.QueueBind(queue, queue, queue, false, nil)
	if err != nil {
		Check(err, "Unable to bind queue")
		return nil, nil, err
	}

	return conn, pub, nil
}

func (c *AMQPConnection) SetupAMQPBroker() {
	var err error
	c.conn, c.pub, err = setup(amqp_uri, amqp_queue)
	CheckPanic(err, "Problem acquiring connection")
}

func (c *AMQPConnection) SendAMQP(parsed Graylog2Parsed) (err error) {
	encoded, err := json.Marshal(parsed)
	Check(err, "Unable to encode json")
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         []byte(encoded),
		Expiration:   amqp_message_ttl,
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.pub != nil {
		err = c.pub.Publish(c.Queue, c.Queue, false, false, msg)
		if err != nil {
			Check(err, "Unable to publish message")
		}
	}

	return err
}

func (c *AMQPConnection) notify_and_reconnect() (err error) {
	bc := make(chan *amqp.Error)
	c.conn.NotifyClose(bc)
	b := <-bc
	for {
		if c.conn == nil || b != nil {
			c.conn, c.pub, err = setup(c.Uri, c.Queue)
			if err == nil {
				return
			}
			time.Sleep(2 * time.Second)
		}
	}
}

func (c *AMQPConnection) ReconnectOnClose() (err error) {
	for {
		c.notify_and_reconnect()
	}
}

func (c *AMQPConnection) Close() {
	defer c.conn.Close()
}
