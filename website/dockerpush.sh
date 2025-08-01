name="website"

GOOS=linux GOARCH=amd64 go build -v -ldflags="-w -s" -o ./$name
docker pull --platform=linux/amd64 alpine:latest
docker build --platform=linux/amd64 --push -t docker.io/injoyai/website-amd64:latest -f ./Dockerfile .

GOOS=linux GOARCH=arm64 go build -v -ldflags="-w -s" -o ./$name
docker pull --platform=linux/arm64 alpine:latest
docker build --platform=linux/arm64 --push -t docker.io/injoyai/website-arm64:latest -f ./Dockerfile .

GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags="-w -s" -o ./$name
docker pull --platform=linux/arm alpine:latest
docker build --platform=linux/arm --push -t docker.io/injoyai/website-arm:latest -f ./Dockerfile .

sleep 8