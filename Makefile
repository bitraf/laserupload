GOSRC = main.go

all: laserupload laserupload.exe

gen_bindata.go: Makefile gen/bindata.go bindata/*
	go run gen/bindata.go

laserupload: Makefile $(GOSRC) gen_bindata.go
	go build

laserupload.exe: Makefile $(GOSRC) gen_bindata.go
	GOOS=windows go build

.PHONY: clean

clean:
	$(RM) laserupload laserupload.exe gen_bindata.go
