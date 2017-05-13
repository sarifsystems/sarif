package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
)

type Watcher struct {
	*app
}

func (w *Watcher) Watch() {
	w.Info.Read()

	err := w.Info.Read()
	if err == nil {
		if w.Info.PID > 0 {
			var p *os.Process
			p, err = os.FindProcess(w.Info.PID)
			if err == nil {
				err = p.Signal(syscall.Signal(0))
			}
		} else {
			err = errors.New("PID not found")
		}
	}

	if err != nil {
		w.Fork()
		return
	}

	mode := ""
	if *isTmux {
		mode = "tmux"
	}

	keys := make([]string, 0, len(w.Info.Bars))
	if len(flag.Args()) > 0 {
		keys = flag.Args()
	} else {
		for key := range w.Info.Bars {
			keys = append(keys, key)
		}
		sort.Strings(keys)
	}

	texts := make([]string, 0)
	for _, key := range keys {
		bar, ok := w.Info.Bars[key]
		if !ok {
			continue
		}
		text := bar.PrettyString(mode)
		if text != "" {
			texts = append(texts, text)
		}
	}
	fmt.Println(strings.Join(texts, *delimiter))
}

func (w *Watcher) Fork() {
	fmt.Println("booting up ...")

	cmd := exec.Command(os.Args[0], "relay")
	stdout, err := cmd.StdoutPipe()
	w.Must(err)
	w.Must(cmd.Start())

	out := bufio.NewReader(stdout)
	out.ReadString('\n')
}
