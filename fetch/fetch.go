package fetch

import (
	"log"
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

func (w Wget) Fetch(uri, gdriveid, filename string, echo bool) error {
	args := []string{"--header=Cookie: gdriveid=" + gdriveid, uri, "-O", filename}
	args = append(args, w.options...)
	cmd := exec.Command("wget", args...)
	if echo {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Start()
	if err != nil {
		log.Println(err)
	}
	return cmd.Wait()
}

func (w *Wget) SetOptions(options ...string) {
	w.options = append(w.options, options...)
}

func (c Curl) Fetch(uri, gdriveid, filename string, echo bool) error {
	args := []string{"-L", uri, "--cookie", "gdriveid=" + gdriveid, "--output", filename}
	args = append(args, c.options...)
	cmd := exec.Command("curl", args...)
	if echo {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Start()
	if err != nil {
		log.Println(err)
	}
	return cmd.Wait()
}

func (c *Curl) SetOptions(options ...string) {
	c.options = append(c.options, options...)
}

func (a Aria2) Fetch(uri, gdriveid, filename string, echo bool) error {
	args := []string{"--header=Cookie: gdriveid=" + gdriveid, uri, "--out", filename}
	args = append(args, a.options...)
	cmd := exec.Command("aria2c", args...)
	if echo {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Start()
	if err != nil {
		log.Println(err)
	}
	return cmd.Wait()
}

func (a *Aria2) SetOptions(options ...string) {
	a.options = append(a.options, options...)
}

func (a Axel) Fetch(uri, gdriveid, filename string, echo bool) error {
	args := []string{"--header=Cookie: gdriveid=" + gdriveid, uri, "--output", filename}
	args = append(args, a.options...)
	cmd := exec.Command("axel", args...)
	if echo {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Start()
	if err != nil {
		log.Println(err)
	}
	return cmd.Wait()
}

func (a *Axel) SetOptions(options ...string) {
	a.options = append(a.options, options...)
}
