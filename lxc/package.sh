#!/usr/bin/env bash
# this script builds `lxc` for OS X, Linux amd64, Windows x64, and Linux arm.
# this script works on "darwin/amd64" and "linux/amd64";

OS=$(uname -s)
if [[ ${OS} == 'Darwin' ]]; then
  this_dir=$(cd $(dirname $0); pwd)
else
  this_dir=$(dirname $(readlink -f $0))
fi

FILES=$(find ${this_dir}/.. -type f -name "*.go" | grep -v "_test.go")
REPOS=$(
for file in ${FILES}; do
  cat ${file} | grep "github.com" | cut -d ' ' -f 1
  cat ${file} | grep "golang.org/x"
done | cut -d '/' -f 1,2,3 | cut -d '"' -f 2 | sort | uniq | grep '/'
)

HASHES=\{$(
for repo in ${REPOS}; do
  pushd ${GOPATH}/src/${repo} >/dev/null
  echo \"${repo}\":\"$(git show -s --format=%H,%cI)\"
  popd >/dev/null
done
)\}

HASHES=$(echo ${HASHES} | sed 's/\ /,/g')

LDFLAGS="-X main.hashes=${HASHES}"

if [[ -z $1 ]]; then
  go install -v -ldflags "${LDFLAGS}"
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
    go build -v -ldflags "${LDFLAGS}" -o lxc/lxc-${os}-${arch}
  done

  tar czf lxc.bin-${VER}.tgz lxc/
  rm -fr lxc/
  echo
  echo -e "\x1b[32mPackage Done\x1b[0m."
fi
