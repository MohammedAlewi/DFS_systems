package file_IO

import (
	"fmt"
	"os"
	"strings"
)


type Message struct {
	FileName   string
	Data []byte
	SIZE int
}


func check_error(err error){
	if err!=nil{
		panic(err)
	}
}

func readFile(filename string) Message{
	file, err := os.Open(filename)
	fmt.Print(err)
	check_error(err)

	fileInfo,_:=file.Stat();
	size:=fileInfo.Size();
	data:=make([]byte,size)

	length,er2:=file.Read(data)
	check_error(er2)

	file_name_path:=strings.Split(filename,"/")
	filename=file_name_path[len(file_name_path)-1]

	msg:=Message{FileName:filename,Data:data,SIZE:length}
	return msg
}


func writeFile(mssgObj Message) bool{
	file_name:="NEW_"+mssgObj.FileName

	file, err := os.Create(file_name)
	check_error(err)

	_,err=file.Write(mssgObj.Data)
	check_error(err)
	return true
}


