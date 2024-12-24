name="forward_linux_amd64"
GOOS=linux GOARCH=amd64 go build -v -ldflags="-w -s" -o ./bin/$name
echo "$name编译完成..."
echo "开始压缩..."
upx -9 -k "./bin/$name"
if [ -f "./bin/$name.~" ]; then
  rm "./bin/$name.~"
fi
if [ -f "./bin/$name.000" ]; then
  rm "./bin/$name.000"
fi

name="forward_linux_arm"
GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags="-w -s" -o ./bin/$name
echo "$name编译完成..."
echo "开始压缩..."
upx -9 -k "./bin/$name"
if [ -f "./bin/$name.~" ]; then
  rm "./bin/$name.~"
fi
if [ -f "./bin/$name.000" ]; then
  rm "./bin/$name.000"
fi

name="forward_windows_amd64"
GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s" -o ./bin/$name.exe
echo "Windows编译完成..."
echo "开始压缩..."
upx -9 -k "./bin/$name.exe"
if [ -f "./bin/$name.ex~" ]; then
  rm "./bin/$name.ex~"
fi
if [ -f "./bin/$name.000" ]; then
  rm "./bin/$name.000"
fi

sleep 8
