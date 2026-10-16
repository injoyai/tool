name="forward"
GOOS=windows GOARCH=amd64 go build -v -ldflags="-H windowsgui -w -s" -o ./$name.exe
echo "$name 编译完成..."
upx -9 -k "./$name.exe"
rm "./$name.ex~"
rm "./$name.000"
echo "压缩完成..."
cmd.exe /c "i upload minio forward.exe"
sleep 2