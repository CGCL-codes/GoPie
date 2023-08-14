GOROOT=$(go env GOROOT)
SRCPATH=$GOROOT/src

cp ./patch/runtime/chan.go.patch $SRCPATH/runtime/chan.go
cp ./patch/sync/mutex.go.patch $SRCPATH/sync/mutex.go
cp ./patch/sync/rwmutex.go.patch $SRCPATH/sync/rwmutex.go
cp ./patch/runtime/runtime2.go.patch $SRCPATH/runtime/runtime2.go
cp ./patch/time/sleep.go.patch $SRCPATH/time/sleep.go
cp -r ./pkg/sched $SRCPATH

