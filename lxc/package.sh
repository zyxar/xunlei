#!/usr/bin/env bash
# this script builds `lxc` for OS X, Linux amd64, Windows x64, Windows i386, and Linux arm.
# this script works on "darwin/amd64" and "linux/amd64";

# retrieve version
VER="$(git branch | grep \* | cut -d ' ' -f 2)-$(git show-ref --head | head -1 | fold -w 8 | head -1)"
sed -i.origin "s/\(version = \)\".*\"/\1\"$VER\"/g" conf.go

if [[ -z $1 ]]; then
  go install -v
else
  mkdir -p lxc

  TARGETLIST="darwin/amd64 linux/amd64 windows/amd64 windows/386 linux/arm"

  # native
  HOSTOS=$(go env GOHOSTOS)
  HOSTARCH=$(go env GOHOSTARCH)
  go build -v -o lxc/lxc-${HOSTOS}-${HOSTARCH}

  # other platforms
  gocross=$(which gocross.py | cut -d ' ' -f 1)
  for target in ${TARGETLIST};do
    os=$(echo ${target} | cut -d '/' -f 1)
    arch=$(echo ${target} | cut -d '/' -f 2)
    suffix=""
    if [[ ${os} == "windows" ]];then
      suffix=".exe"
    fi
    if [[ ${os} != ${HOSTOS} || ${arch} != ${HOSTARCH} ]];then
      (test ! -z ${gocross}) && (test -f ${gocross}) && (${gocross} ${target} build -v -o lxc/lxc-${os}-${arch}${suffix})
    fi
  done

  tar czf lxc.bin-${VER}.tgz lxc/
  rm -fr lxc/
fi
# restore version
mv conf.go.origin conf.go

echo
echo -e "\x1b[32mPackage Done\x1b[0m."