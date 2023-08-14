#!/bin/bash
if [ "$1" = "go" ]
then
  mv ./patch/runtime/chan.go.patch ./patch/runtime/chan.go
  mv ./patch/sync/mutex.go.patch ./patch/sync/mutex.go
  mv ./patch/sync/rwmutex.go.patch ./patch/sync/rwmutex.go
  mv ./patch/runtime/runtime2.go.patch ./patch/runtime/runtime2.go
else
  mv ./patch/runtime/chan.go ./patch/runtime/chan.go.patch
  mv ./patch/sync/mutex.go ./patch/sync/mutex.go.patch
  mv ./patch/sync/rwmutex.go ./patch/sync/rwmutex.go.patch
  mv ./patch/runtime/runtime2.go ./patch/runtime/runtime2.go.patch
fi