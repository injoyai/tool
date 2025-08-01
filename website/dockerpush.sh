name="website"

docker pull --platform=linux/amd64 alpine:latest

GOOS=linux GOARCH=amd64 go build -v -ldflags="-w -s" -o ./$name
upx -9 -k "./$name"
rm "./$name.~"
rm "./$name.000"
docker build --platform=linux/amd64 --push -t docker.io/injoyai/website-amd64:latest -f ./Dockerfile .

GOOS=linux GOARCH=arm64 go build -v -ldflags="-w -s" -o ./$name
upx -9 -k "./$name"
rm "./$name.~"
rm "./$name.000"
docker build --platform=linux/arm64 --push -t docker.io/injoyai/website-arm64:latest -f ./Dockerfile .

GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags="-w -s" -o ./$name
upx -9 -k "./$name"
rm "./$name.~"
rm "./$name.000"
docker build --platform=linux/arm7 --push -t docker.io/injoyai/website-arm7:latest -f ./Dockerfile .

sleep 8