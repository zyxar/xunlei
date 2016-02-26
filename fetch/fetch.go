package fetch

import (
	"os"
	"os/exec"
)

type Fetcher interface {
	Fetch(uri, gdriveid, filename string, echo bool) error
	SetOptions(options ...string)
}

type Wget struct {
	options []string
}

type Curl struct {
	options []string
}

type Aria2 struct {
	options []string
}

type Axel struct {
	options []string
}

var DefaultFetcher Fetcher

func init() {
	DefaultFetcher = &Aria2{
		options: []string{"--file-allocation=none", "-s5", "-x5", "-c", "--summary-interval=0", "--follow-torrent=false"},
	}
}

func callCommand(cmd string, args []string, echo bool) error {
	command := exec.Command(cmd, args...)
	if echo {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}
	if err := command.Start(); err != nil {
		return err
	}
	return command.Wait()
}

func (w Wget) Fetch(uri, gdriveid, filename string, echo bool) error {
	args := []string{"--header=Cookie: gdriveid=" + gdriveid, uri, "-O", filename}
	args = append(args, w.options...)
	return callCommand("wget", args, echo)
}

func (w *Wget) SetOptions(options ...string) {
	w.options = append(w.options, options...)
}

func (c Curl) Fetch(uri, gdriveid, filename string, echo bool) error {
	args := []string{"-L", uri, "--cookie", "gdriveid=" + gdriveid, "--output", filename}
	args = append(args, c.options...)
	return callCommand("curl", args, echo)
}

func (c *Curl) SetOptions(options ...string) {
	c.options = append(c.options, options...)
}

func (a Aria2) Fetch(uri, gdriveid, filename string, echo bool) error {
	args := []string{"--header=Cookie: gdriveid=" + gdriveid, uri, "--out", filename}
	args = append(args, a.options...)
	return callCommand("aria2c", args, echo)
}

func (a *Aria2) SetOptions(options ...string) {
	a.options = append(a.options, options...)
}

func (a Axel) Fetch(uri, gdriveid, filename string, echo bool) error {
	args := []string{"--header=Cookie: gdriveid=" + gdriveid, uri, "--output", filename}
	args = append(args, a.options...)
	return callCommand("axel", args, echo)
}

func (a *Axel) SetOptions(options ...string) {
	a.options = append(a.options, options...)
}
