set GOOS=windows
go build -o ./out/wsgo.exe main.go  
set GOARCH=amd64
set GOOS=linux
go build -o ./out/wsgo main.go



