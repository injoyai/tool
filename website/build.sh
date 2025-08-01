name="website"

GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s" -o ./$name.exe
echo "$name 编译完成..."
echo "开始压缩..."
upx -9 -k "./$name.exe"
rm "./$name.ex~"
rm "./$name.000"

sleep 8
