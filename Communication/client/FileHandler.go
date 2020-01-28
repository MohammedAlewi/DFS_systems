package main

import (
	"fmt"
	"hash/fnv"
	"os"
	"strings"
)

type FileHandler struct {
	FileName   string
	Data []byte
	SIZE int
	filenameHash string
}


func (message *FileHandler) setFileHandler(msgObj Message){
	message.FileName=msgObj.FileName
	message.Data=msgObj.Data
	message.SIZE=msgObj.SIZE
}

func (message *FileHandler) readFile() Message{
	file, err := os.Open(message.FileName)
	message.checkError(err)

	fileInfo,_:=file.Stat();
	size:=fileInfo.Size();
	data:=make([]byte,size)

	length,er2:=file.Read(data)
	message.checkError(er2)

	fileNamePath :=strings.Split(message.FileName,"/")
	message.FileName= fileNamePath[len(fileNamePath)-1]
	message.Data=data
	message.SIZE=length
	msg:= Message{FileName:message.FileName,Data:message.Data,SIZE:message.SIZE}
	return msg
}


func (message *FileHandler) writeFile(dirPath string) bool{
	fileName :=dirPath+"/"+message.hash()+"__"+message.FileName

	file, err := os.Create(fileName)
	message.checkError(err)
	_,err=file.Write(message.Data[:len(message.Data)])
	message.checkError(err)

	file.Close()
	return true
}


func (message *FileHandler)  hash() string {
	h := fnv.New32a()
	h.Write([]byte(message.FileName))
	message.filenameHash=fmt.Sprint(h.Sum32())
	return message.filenameHash
}


func(message *FileHandler) checkError(err error){
	if err!=nil{
		panic(err)
	}
}

