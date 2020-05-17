@echo off&setlocal

echo Building docker container ...
docker build -q -t gmcbuilder -f ./build/Dockerfile_gmcbuilder build 
docker run -it --rm --mount type=bind,source="%CD%",target=/app -w /app gmcbuilder go run /app/build/build.go

pause