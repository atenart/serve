/*
Copyright (C) 2015-2017 Antoine Tenart

Antoine Tenart <antoine.tenart@ack.tf>

This file is licensed under the terms of the GNU General Public License version
2.  This program is licensed "as is" without any warranty of any kind, whether
express or implied.
*/

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/atotto/clipboard"
	docopt "github.com/docopt/docopt-go"
)

var (
	version = "serve 0.2"
	usage   = `Temporary HTTP server to share local files

Usage:
  serve [options] [FILE]
  serve --help

Options:
  FILE                                 File or directory to serve [default: .]
  -p <port>, --port <port>             Port number to listen on [default: 8080]
  -c <count>, --count <count>          Limit the number of allowed GET to <count>
  -h --help                            Print this help
`
)

func outboundIP() net.IP {
	c, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	return c.LocalAddr().(*net.UDPAddr).IP
}

func main() {
	var handler http.Handler = nil
	var count int = -1

	args, _ := docopt.Parse(usage, os.Args[1:], true, version, true)

	f, _ := args["FILE"].(string)
	f = path.Clean(f)
	if f == "" {
		f = "."
	}

	resource := path.Base(f)

	fi, err := os.Stat(f)
	if err != nil {
		log.Fatal(err)
	}

	switch mode := fi.Mode(); {
	case mode.IsRegular():
		if limit, ok := args["--count"].(string); ok {
			var err error

			count, err = strconv.Atoi(limit)
			if err != nil {
				log.Fatal(err)
			}
		}

		if resource == "." {
			resource = ""
		}

		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var extra string = ""
			if count > 1 {
				extra = fmt.Sprintf(", %d remaining GET",
					count-1)
			}

			client, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				client = "unknown client"
			}
			fmt.Printf("Connexion from %s requesting %s%s\n", client,
				r.URL.Path, extra)

			if r.URL.Path != fmt.Sprintf("/%s", resource) {
				http.Error(w, "File not found", 404)
				return
			}
			if count > 0 {
				count--
			}

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Transfer-Encoding", "binary")
			w.Header().Set("Content-Disposition",
				fmt.Sprintf("attachment; filename=%s", resource))
			w.Header().Set("Content-Length",
				strconv.FormatInt(fi.Size(), 10))
			w.Header().Set("Cache-Control", "private")
			w.Header().Set("Pragma", "private")
			w.Header().Set("Expires", "0")

			http.ServeFile(w, r, f)
			if count == 0 {
				os.Exit(0)
			}
		})
	case mode.IsDir():
		resource = ""
		handler = http.FileServer(http.Dir(f))
	default:
		log.Fatal("Error: unsupported file type.")
	}

	uri := fmt.Sprintf("http://%s:%s/%s", outboundIP(), args["--port"].(string), resource)
	fmt.Printf("Serving %s at %s\n", f, uri)

	err = clipboard.WriteAll(uri)
	if err != nil {
		log.Print(err)
	}

	err = http.ListenAndServe(":"+args["--port"].(string), handler)
	if err != nil {
		log.Fatal(err)
	}
}
