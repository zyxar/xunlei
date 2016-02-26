//+build windows

package main

import (
	"os"
	"path"

	"golang.org/x/crypto/ssh/terminal"
)

func initHome() {
	home = path.Dir(os.Args[0])
}

func clearscr() {
}

type wterm struct {
	t *terminal.Terminal
}

func (*wterm) Read(b []byte) (int, error) {
	return os.Stdin.Read(b)
}

func (*wterm) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}

func newTerm() Term {
	w := new(wterm)
	w.t = terminal.NewTerminal(w, "lixian >> ")
	return w
}

func (w *wterm) ReadLine() (string, error) {
	return w.t.ReadLine()
}

func (w *wterm) Restore() {
}
