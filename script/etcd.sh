for gomod in `find . -name "go.mod"`
do
  sed -i "s/go 1.[0-9]\+/go 1.18/g" $gomod
  echo "replace google.golang.org/grpc => google.golang.org/grpc v1.29.1" >> $gomod
  cd $(dirname $gomod)
  go mod tidy
  cd /tool
done