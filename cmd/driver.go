/*
 *  Copyright (C) [SonicCloudOrg] Sonic Project
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/SonicCloudOrg/sonic-ios-bridge/src/entity"
	"github.com/SonicCloudOrg/sonic-ios-bridge/src/util"
	"github.com/electricbubble/gwda"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"log"
	"net"
	"net/http"
)

var driverCmd = &cobra.Command{
	Use:   "driver",
	Short: "Start iOSDriver in webSocket.",
	Long:  "Start iOSDriver in webSocket.",
	RunE: func(cmd *cobra.Command, args []string) error {
		//Get free port.
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			return util.NewErrorPrint(util.ErrUnknown, "Get free port.", err)
		}
		var port = listener.Addr().(*net.TCPAddr).Port
		log.Println("Using port:", port)
		listener.Close()
		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()
		r.GET("/driver", driver)
		r.Run(fmt.Sprintf(":%d", port))
		return nil
	},
}

var remoteUrl string

func init() {
	rootCmd.AddCommand(driverCmd)
	driverCmd.Flags().StringVarP(&remoteUrl, "remote-url", "r", "http://localhost:8100", "device's wda remote url (default http://localhost:8100)")
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func driver(c *gin.Context) {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	log.Println("onOpened.")

	driver, err := gwda.NewDriver(nil, remoteUrl)
	if err != nil {
		log.Println("init driver failed! cause: " + err.Error())
		return
	}
	healthy, err := driver.IsWdaHealthy()
	if err != nil || !healthy {
		log.Println("wda not health!")
		return
	}
	defer driver.Close()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		s := &entity.WebSocketReq{}
		json.Unmarshal(message, s)
		log.Println(fmt.Sprintf("method:%s,params:%s", s.Method, s.Params))
		if s.Version == "1.4.2" {
			wr := &entity.WebSocketRep{}
			switch s.Method {
			case "source":
				var r string
				r, err = driver.Source()
				if err != nil {
					log.Println(err)
				}
				wr.Result = r
			//case "send":
			//	driver
			case "tap":
				err = driver.Tap(s.Params[0].(int), s.Params[1].(int))
				if err != nil {
					log.Println(err)
				}
				wr.Result = "done"
			case "swipe":
				err = driver.Swipe(s.Params[0].(int), s.Params[1].(int), s.Params[2].(int), s.Params[3].(int))
				if err != nil {
					log.Println(err)
				}
				wr.Result = "done"
			}
			wr.Id = s.Id
		}
	}
}
