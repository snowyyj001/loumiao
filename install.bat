@echo off
setlocal
if exist install.bat goto ok
echo install.bat must be run from its folder
goto end
: ok
set OLDGOPATH=%GOPATH%
set GOPATH=%~dp0

echo gopath: %GOPATH%
gofmt -w src

::uuid
echo begin install uuid module
go get github.com/google/uuid

::mysql
echo begin install mysql module
go get -u github.com/go-sql-driver/mysql

::mongo
echo begin install mongo module
go get gopkg.in/mgo.v2

::redis
echo begin install redis module
go get github.com/gomodule/redigo/redis

::gorm
echo begin install gorm module
go get -u github.com/jinzhu/gorm

::websocket
echo begin install websocket module
go get github.com/gorilla/websocket

::log
echo begin install logger module
go get github.com/phachon/go-logger

::go install loumiao

:end
echo finished

pause
