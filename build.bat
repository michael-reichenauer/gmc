@echo off&setlocal

del gmc.exe >nul 2>&1
del gmc_linux >nul 2>&1
del gmc_darwin >nul 2>&1

rem "-s -w" omits debug and symbols to reduse size ("-H=windowsgui" would disable console on windows)

echo Building gmc.exe ...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o gmc.exe main.go

echo Building gmc_linux ...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o gmc_linux main.go

echo Building gmc_darwin ...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-s -w" -o gmc_darwin main.go


echo.
echo Built gmc binaries:
.\gmc.exe -version

echo.
echo.
pause