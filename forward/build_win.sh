name="forward"
GOOS=windows GOARCH=amd64 go build -v -ldflags="-H windowsgui -w -s" -o ./bin/$name.exe
echo "$name 编译完成..."
echo "开始压缩..."
upx -9 -k "./bin/$name.exe"
if [ -f "./bin/$name.ex~" ]; then
  rm "./bin/$name.ex~"
fi
if [ -f "./bin/$name.000" ]; then
  rm "./bin/$name.000"
fi

echo "上传至minio"
cmd.exe /c "in upload minio ./bin/$name.exe"

sleep 8