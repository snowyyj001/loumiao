#!/usr/bin bash
if [ ! -f install ]; then
	echo 'install must be run within its container folder' 1>&2
	exit 1
fi
CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR"
echo gopath: "$GOPATH"
gofmt -w src

export GOPATH="$OLDGOPATH"

#uuid
echo begin install uuid module
go get github.com/google/uuid

#mysql
echo begin install mysql module
go get -u github.com/go-sql-driver/mysql

#mongo
echo begin install mongo module
go get gopkg.in/mgo.v2

#redis
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

echo 'finished'
