cd /tmp
wget https://api.gogen.io/osx/gogen
chmod 755 ./gogen
./gogen -c coccyx/weblog
./gogen search weblog
./gogen pull -d coccyx/weblog .
less samples/weblog.json
./gogen info coccyx/weblog
./gogen -c coccyx/weblog gen -c 1 -ei 1
./gogen -c coccyx/weblog gen -c 1000 -ei 1000 -o devnull

gogen -g 3 -c examples/generator/users.yml gen -ei 10000