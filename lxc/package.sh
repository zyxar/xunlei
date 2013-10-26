#!/usr/bin/env bash
# this script build lxc for OS X, Linux amd64, Windows x64, Windows i386, and Linux arm.
# this script is supposed to run on osx


# retrieve version
VER="$(git branch | grep \* | cut -d ' ' -f 2)-$(git show-ref --head | head -1 | fold -w 8 | head -1)"
sed -i.origin "s/\(version = \)\".*\"/\1\"$VER\"/g" conf.go

if [[ -z $1 ]]; then
  go install -v
else
  mkdir -p lxc
  # osx
  go build -v -o lxc/lxc-osx

  # other platforms
  gocross=$(which gocross.py | cut -d ' ' -f 1)
  (test ! -z ${gocross}) && (test -f ${gocross}) && (${gocross} win64 build -v -o lxc/lxc-win64.exe)
  (test ! -z ${gocross}) && (test -f ${gocross}) && (${gocross} win32 build -v -o lxc/lxc-win32.exe)
  (test ! -z ${gocross}) && (test -f ${gocross}) && (${gocross} linux build -v -o lxc/lxc-linux64)
  (test ! -z ${gocross}) && (test -f ${gocross}) && (${gocross} arm build -v -o lxc/lxc-arm)

  tar czf lxc.bin.tgz lxc
  rm -fr lxc/
fi
# restore version
mv conf.go.origin conf.go
