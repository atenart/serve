/*
 * Copyright (C) 2015 Antoine Tenart
 *
 * Antoine Tenart <antoine.tenart@ack.tf>
 *
 * This file is licensed under the terms of the GNU General Public
 * License version 2.  This program is licensed "as is" without any
 * warranty of any kind, whether express or implied.
 */

package main

import (
	"errors"
	"net/http"
	"os"
	"path"

	docopt "github.com/docopt/docopt-go"
)

var (
	version = "webserv 0.1"
	usage = `Temporary http server to serve files

Usage:
  webserv [-d <directory> | -f <file>] [-p <port>]
  webserv --help

Options:
  -d <directory>, --dir <directory>    Serve the given directory [default: ./]
  -f <file>, --file <file>             Serve the given file
  -p <port>, --port <port>             Port number to listen on [default: 8080]
  --help                               Print this help
`
)

type File string

func (f File) Open(name string) (http.File, error) {
	if name != ("/" + path.Clean(string(f))) {
		return nil, errors.New("http: invalid request")
	}

	file, err := os.Open(string(f))
	if err != nil {
		return nil, err
	}

	return file, nil
}

func main() {
	var fs http.FileSystem

	args, _ := docopt.Parse(usage, os.Args[1:], true, version, true)

	if f, ok := args["--file"].(string); ok {
		fs = File(f)
	} else {
		fs = http.Dir(args["--dir"].(string))
	}
	http.Handle("/", http.FileServer(fs))
	http.ListenAndServe(":" + args["--port"].(string), nil)
}
