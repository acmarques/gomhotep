all: install

install:
	go build -ldflags '-linkmode external -extldflags "-static -lm -lz -lbz2 -lltdl -pthread -ldl -lclamunrar"' gomhotep.go
