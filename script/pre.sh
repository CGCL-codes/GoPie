cd /tool

rm -r ./testbins
PWD=`pwd`
cd $1 || exit
git checkout -- *
find . -name "go.mod" | sed -i "s/go 1.[0-9]\+/go 1.18/g" ./go.mod
cd /tool || exit

go build -o ./bin ./cmd/...
./bin/fuzz --task inst --path $1
# shellcheck disable=SC2086
cd $1 || exit
# shellcheck disable=SC2086
cd /tool || exit

./script/patch.sh
./bin/fuzz --task bins --path $1