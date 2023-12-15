#!/bin/bash

DIR="$(cd $(dirname ${BASH_SOURCE:-$0})/../; pwd)"
cd "$DIR" || exit 1

ESC=$(printf '\033')
FAILED=0

run-test() {
  IMAGE_NAME="${1%:*}"
  FILE_NAME="${IMAGE_NAME/\//_}"
  TOOL="$2"

  PREPARE_CMD="$3"
  if [[ "$PREPARE_CMD" != "" ]]; then
    PREPARE_CMD="$PREPARE_CMD; "
  fi

  CMD="${PREPARE_CMD}spirat -tools=$TOOL -force -filename /tmp/output/$FILE_NAME.json"
  echo "Testing with: $CMD"

  mkdir -p test/output
  touch "test/output/$FILE_NAME.json"

  docker run --rm \
    -v "$PWD/build/spirat:/usr/bin/spirat" \
    -v "$PWD/test/output:/tmp/output" \
    "$1" bash -c "$CMD" \
    > "test/output/$FILE_NAME.log" \
    2> "test/output/$FILE_NAME.err.log"
  SUCCESS=$?

  if [[ $SUCCESS = 0 ]]; then
    echo "OK: $1"
  else
    echo "${ESC}[31mFAILED${ESC}[m: $1"
    FAILED=1
  fi
}

make

run-test almalinux:8.4            rpm
run-test amazonlinux:2            rpm
run-test centos:8                 rpm
run-test debian:11.1              dpkg
run-test opensuse/leap            rpm
run-test oraclelinux:8.4          rpm
run-test ubuntu:20.04             dpkg
run-test node:20-bullseye-slim    npm   "cd; npm install react"

if [[ $FAILED = 1 ]]; then
  echo
  echo "${ESC}[31mTest failed${ESC}[m"
  exit 1
fi
