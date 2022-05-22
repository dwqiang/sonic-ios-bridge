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
	"fmt"
	"github.com/SonicCloudOrg/sonic-ios-bridge/src/util"
	giDevice "github.com/electricbubble/gidevice"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Proxy unix process or port to your pc.",
	Long:  "Proxy unix process or port to your pc.",
	RunE: func(cmd *cobra.Command, args []string) error {
		usbMuxClient, err := giDevice.NewUsbmux()
		if err != nil {
			return util.NewErrorPrint(util.ErrConnect, "usbMux", err)
		}
		list, err1 := usbMuxClient.Devices()
		if err1 != nil {
			return util.NewErrorPrint(util.ErrSendCommand, "listDevices", err1)
		}
		if len(list) == 0 {
			fmt.Println("no device connected")
			os.Exit(0)
		} else {
			var device giDevice.Device
			if len(udid) != 0 {
				for i, d := range list {
					if d.Properties().SerialNumber == udid {
						device = list[i]
						break
					}
				}
			} else {
				device = list[0]
			}
			serverListener, err := net.Listen("tcp", fmt.Sprintf(":%d", localPort))
			if err != nil {
				return err
			}
			defer serverListener.Close()
			go func(listener net.Listener) {
				for {
					var accept net.Conn
					var err error
					if accept, err = listener.Accept(); err != nil {
						log.Println("accept:", err)
					}
					fmt.Println("accept", accept.RemoteAddr())
					rInnerConn, err := device.NewConnect(remotePort)
					if err != nil {
						fmt.Println("connect to device fail")
						os.Exit(0)
					}
					rConn := rInnerConn.RawConn()
					_ = rConn.SetDeadline(time.Time{})
					go func(lConn net.Conn) {
						go func(lConn, rConn net.Conn) {
							if _, err := io.Copy(lConn, rConn); err != nil {
								log.Println("local -> remote failed:", err)
							}
						}(lConn, rConn)
						go func(lConn, rConn net.Conn) {
							if _, err := io.Copy(rConn, lConn); err != nil {
								log.Println("local <- remote failed:", err)
							}
						}(lConn, rConn)
					}(accept)
				}
			}(serverListener)
		}
		return nil
	},
}

var remotePort, localPort int

func init() {
	rootCmd.AddCommand(proxyCmd)
	proxyCmd.Flags().StringVarP(&udid, "udid", "u", "", "device's serialNumber ( default first device )")
	proxyCmd.Flags().IntVarP(&remotePort, "remote-port", "r", 8100, "remote port")
	proxyCmd.Flags().IntVarP(&localPort, "local-port", "l", 9100, "local port")
}
