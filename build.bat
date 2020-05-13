@echo off&setlocal

echo Running tests ...
go test ./...
if %ERRORLEVEL% neq 0 goto :ERROR

del gmc.exe >nul 2>&1
del gmc_linux >nul 2>&1
del gmc_mac >nul 2>&1

rem "-s -w" omits debug and symbols to reduce size ("-H=windowsgui" would disable console on windows)

echo Building gmc.exe ...
set GOOS=windows
set GOARCH=amd64
go build -tags release -ldflags "-s -w" -o gmc.exe main.go
if %ERRORLEVEL% neq 0 goto :ERROR

echo Building gmc_linux ...
set GOOS=linux
set GOARCH=amd64
go build -tags release -ldflags "-s -w" -o gmc_linux main.go
if %ERRORLEVEL% neq 0 goto :ERROR

echo Building gmc_mac ...
set GOOS=darwin
set GOARCH=amd64
go build -tags release -ldflags "-s -w" -o gmc_mac main.go
if %ERRORLEVEL% neq 0 goto :ERROR


echo.
echo Built gmc binaries:
.\gmc.exe -version
if %ERRORLEVEL% neq 0 goto :ERROR

echo.
echo.
pause

EXIT /B %errorlevel%

:ERROR
echo.
echo ERROR has occurred !!
pause