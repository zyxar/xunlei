#!/usr/bin/env bash
# this script builds `lxc` for OS X, Linux amd64, Windows x64, and Linux arm.
# this script works on "darwin/amd64" and "linux/amd64";

LDFLAGS="-X main.version=$(git rev-parse HEAD | fold -w 8 | head -1)"

if [[ -z $1 ]]; then
  go build -v -ldflags ${LDFLAGS}
else
  TARGETLIST="darwin/amd64 linux/amd64 windows/amd64 linux/arm"
  mkdir -p lxc

  for target in ${TARGETLIST};do
    os=$(echo ${target} | cut -d '/' -f 1)
    arch=$(echo ${target} | cut -d '/' -f 2)
    suffix=""
    if [[ ${os} == "windows" ]];then
      suffix=".exe"
    fi
    go build -v -ldflags ${LDFLAGS} -o lxc/lxc-${os}-${arch}
  done

  tar czf lxc.bin-${VER}.tgz lxc/
  rm -fr lxc/
fi

echo
echo -e "\x1b[32mPackage Done\x1b[0m."