#!/usr/bin/env bash

set -e

rootpath=$(dirname $(readlink -f $0))/..
imageVersion=$(date +%Y%m%d)
dockerDir=${rootpath}/build/docker
buildDir=${rootpath}/cmd

dockerHub=${DockerHub:-"10.5.26.86:8080"}

cleanBinary() {
  rm -fr "${buildDir}"/imanager
}

buildBinary() {
  pushd "${buildDir}"
  echo "building binary..."
  CGO_ENABLED=0 go build -o imanager .
  popd
}

cleanDockerDir() {
  rm -fr "${dockerDir}"/imanager
}

copyFilesToDockerDir() {
  cp "${buildDir}"/imanager "${dockerDir}"
}

makeDockerImage() {
  pushd "${dockerDir}"
  echo "make docker image..."
  cleanDockerDir
  copyFilesToDockerDir
  docker build -t zjlab/imanager:"${imageVersion}" .
  cleanDockerDir
  popd
}

pushDockerImage() {
  docker login "${dockerHub}" -u wangxj --password Harbor123456
  docker tag zjlab/imanager:"${imageVersion}" "${dockerHub}"/zjlab/imanager:"${imageVersion}"
  docker push "${dockerHub}"/zjlab/imanager:"${imageVersion}"
  docker logout "${dockerHub}"
}

main() {
  cleanBinary
  buildBinary

  makeDockerImage
  pushDockerImage
  cleanBinary
}

main