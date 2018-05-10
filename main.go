package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/codeskyblue/pk/cmdctrl"
	shellquote "github.com/kballard/go-shellquote"
)

var (
	version  = "dev"
	fDaemon  bool
	fLogfile string

	service *cmdctrl.CommandCtrl
)

func init() {
	flag.BoolVar(&fDaemon, "d", false, "run daemon mode")
	flag.StringVar(&fLogfile, "log", "pk.log", "log file works only in daemon mode")
}

type Process struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	args    []string
}

func loadConfiguration() {
	service = cmdctrl.New()
	defer service.StopAll()

	jsonFd, err := os.Open("pk.json")
	if err != nil {
		log.Printf("pk.json read error: %s", err)
	}
	defer jsonFd.Close()
	var processes []Process
	if err = json.NewDecoder(jsonFd).Decode(&processes); err != nil {
		log.Fatal(err)
	}

	for _, p := range processes {
		p.args, err = shellquote.Split(p.Command)
		if err != nil {
			log.Fatal("%s fail in split command: %s", p.Name, err)
		}
		log.Println(p.Name, p.args)
		service.Add(p.Name, cmdctrl.CommandInfo{
			Args: p.args,
		})
	}
}

func main() {
	fVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *fVersion {
		fmt.Printf("version %s\n", version)
		return
	}

	if fDaemon && os.Getenv(DAEMON_ENV) == "" {
		runDaemon()
		return
	}

	// ignore sighup in daemon mode
	if os.Getenv(DAEMON_ENV) != "" {
		logfd := setupLogfile(fLogfile)
		defer logfd.Close()

		println("enter into daemon mode")
		signal.Ignore(syscall.SIGHUP)
	}

	// Normal logic goes here
	loadConfiguration()

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		err := service.Start(name)
		if err == nil {
			io.WriteString(w, "success")
		} else {
			io.WriteString(w, err.Error())
		}
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		err := service.Stop(name)
		if err == nil {
			io.WriteString(w, "success")
		} else {
			io.WriteString(w, err.Error())
		}
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		pss := service.AllStatus()
		var lines = []string{"NAME\tRUNNING"}
		for _, ps := range pss {
			lines = append(lines, fmt.Sprintf("%s\t%v", ps.Name, ps.Running))
		}
		io.WriteString(w, strings.Join(lines, "\n"))
	})

	http.ListenAndServe(":18510", nil)
}
