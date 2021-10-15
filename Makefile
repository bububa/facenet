all: *

proto: 
	protoc --go_out=./core ./core/*.proto	
facenet:
	go build -o=./bin/facenet ./cmd/facenet
cvcamera:
	go build -o=./bin/cvcamera -tags=cv4 ./cmd/camera
linux_camera:
	go build -o=./bin/linux_camera -tags=linux ./cmd/camera
android_camera:
	go build -o=./bin/android_camera -tags=android ./cmd/camera
