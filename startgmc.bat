set arg1=%1
go build -gcflags="all=-N -l" -o gmc.exe main.go
start gmc.exe

