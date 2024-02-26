VER=2.5.3
GOARCH=amd64

#all: clean darwin linux windows
all: clean darwin

darwin:
	echo "Make darwin binary ..."
	GOOS=darwin GOARCH=${GOARCH} go build -o ~/gows/bin/darwin/cube_darwin_${VER}
	ln -s ~/gows/bin/darwin/cube_darwin_${VER} ~/gows/bin/darwin/cube


linux:
	echo "Make linux binary ..."
	GOOS=linux GOARCH=${GOARCH} go build -o ./bin/Goproxy .

windows:
	echo "Make windows binary ..."
	GOOS=windows GOARCH=${GOARCH} go build -o ~/gows/bin/windows/cube_windows_${VER}.exe
	ln -s ~/gows/bin/windows/cube_windows_${VER}.exe ~/gows/bin/windows/cube.exe


clean:
	rm -f ~/gows/bin/darwin/cube_darwin_${VER}
	rm -f ~/gows/bin/darwin/cube
	rm -f ~/gows/bin/linux/cube_linux_${VER}
	rm -f ~/gows/bin/linux/cube
	rm -f ~/gows/bin/windows/cube_windows_${VER}.exe
	rm -f ~/gows/bin/windows/cube.exe
