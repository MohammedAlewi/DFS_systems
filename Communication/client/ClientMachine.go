package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)



type Client struct {
	serverAddress string
	serverPORT    string
	address       string
	PORT          string
	listener      net.Listener
	wg sync.WaitGroup
	connection net.Conn
	fileHandler FileHandler
	dbHandler  FileEntry
	mainServerDir string
}



func (client *Client) creatConnection() bool{
	client.wg.Add(1)
	conn, err := net.Dial("tcp", client.serverAddress+":"+client.serverPORT)
	if err!=nil{
		fmt.Println("[Client]: Unable to establish Connection with the remote node")
		return false
	}
	client.connection=conn
	fmt.Println("[Client]: Connection established with the remote node")
	return true
}

func (client *Client) creatListener(){
	client.wg.Add(1)

	listener, err1  := net.Listen("tcp", ":"+client.PORT)
	client.checkError(err1)
	client.listener=listener
	fmt.Println("[Client]: Waiting to accept file from the remote node")
	conn,err :=  listener.Accept()
	client.connection=conn
	client.checkError(err)
	go client.recvMessageFromServer()
}

func (client *Client) sendMessageToServer(messageObj Message){
	defer client.wg.Done()

	binBuf := new(bytes.Buffer)
	goBufferObj := gob.NewEncoder(binBuf)
	goBufferObj.Encode(messageObj)

	_,err2:=client.connection.Write(binBuf.Bytes())
	client.connection.Close()
	client.checkError(err2)
	fmt.Printf("[Client]: Message sent to the remote [NODE] with Address %s:%s\n",client.serverAddress,client.serverPORT)
}

func (client *Client) recvMessageFromServer(){
	defer client.wg.Done()

	dec := gob.NewDecoder(client.connection)
	messageObject := &Message{}
	dec.Decode(messageObject)
	client.connection.Close()

	switch messageObject.COMMAND {
		case  FILE_FOUND:{
			client.fileHandler.setFileHandler(*messageObject)
			client.fileHandler.writeFile(client.mainServerDir+"/Received_Files")
			fmt.Printf("[Client]: message have have been accepted from [NODE] with Address %s:%s\n",messageObject.SenderIP,messageObject.SenderPort)
			fmt.Println("[Client]: file ",messageObject.FileName," have been accepted and saved in local working dir")
		}
		case FILE_NOT_FOUND:{
			fmt.Printf("[Client]: message have have been accepted from [NODE] with Address %s:%s\n",messageObject.SenderIP,messageObject.SenderPort)
			fmt.Println("[Client]: remote node says file ",messageObject.FileName," NOT FOUND!!")
		}
	}
	node:=strconv.FormatUint(uint64(messageObject.NodeID),10);
	client.dbHandler.insertRoutingAddress(node,messageObject.SenderIP,messageObject.SenderPort,messageObject.FileName)

}

func (client *Client) createWorkingDir() {
	dir:="Client/Received_Files"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
	client.mainServerDir="Client"
}


func (client *Client) closeListenerConnection(){
	client.wg.Wait()
	client.listener.Close()
}

func (client *Client) closeConnection(){
	client.wg.Wait()
}

func (client *Client) checkError(err error){
	if err!=nil{
		panic(err)
	}
}


// interactions


func (client *Client) sendFileToServer(filename string) bool{
	client.fileHandler.FileName=filename
	message:=client.fileHandler.readFile()
	message.SenderIP=client.address
	message.SenderPort=client.PORT
	message.COMMAND=SAVE_FILE
	result:=client.creatConnection()
	if result{client.sendMessageToServer(message)
	client.closeConnection()
	return true}
	return false
}

func (client *Client) getFileFromServer(filename string){
	message:=Message{}
	message.FileName=filename
	message.SIZE=12
	message.COMMAND=GET_FILE
	message.SenderIP=client.address
	message.SenderPort=client.PORT

	client.creatConnection()
	client.sendMessageToServer(message)
	client.closeConnection()

	client.creatListener()
	client.closeListenerConnection()
}

func (client *Client)  sendToBestLocationForFile(fileName string){
	node:=client.dbHandler.getFileEntries(strings.Split(fileName,"/")[len(strings.Split(fileName,"/"))-1])
	if node{
		fmt.Printf("[Client]: The Exact Location Found which is %s:%s\n",
			client.dbHandler.nodeAddress[0],client.dbHandler.nodeAddress[1])
		client.serverPORT=client.dbHandler.nodeAddress[1]
		client.serverAddress=client.dbHandler.nodeAddress[0]
		result:=client.sendFileToServer(fileName)
		if result{return}
	}
	client.chooseNextBestLocation(fileName)
}

func (client *Client)  queryNextBestLocationForFile(fileName string)map[int][]string {
	node:=client.dbHandler.getAllRoutingID()
	weights:=make(map[int][]string)
	client.fileHandler.FileName=strings.Split(fileName,"/")[len(strings.Split(fileName,"/"))-1]
	fileBinHash,_:=strconv.ParseUint(client.fileHandler.hash(), 10, 32)
	for _,routeID := range node{
		intId,_:=strconv.ParseUint(routeID[0], 10, 32)
		currentWeight:=intId ^ fileBinHash
		weights[int(currentWeight)]= routeID
	}
	return weights
}

func (client *Client) chooseNextBestLocation( fileName string){
	nodes:=client.queryNextBestLocationForFile(fileName)
	//fmt.Println("flaso",nodes)
	keys:=[]int{}
	for  key := range nodes  {
		keys=append(keys, key)
	}
	sort.Ints(keys)
	for key:=0; key < len(keys);key++{
		fmt.Printf("[Client]: Unable to find the Exact Location. choosing the best location.\n")
		nodeAddress:=nodes[keys[key]]
		client.serverAddress=nodeAddress[1]
		client.serverPORT=nodeAddress[2]
		result := client.sendFileToServer(fileName)
		if result {break}
		fmt.Printf("[Client]: Unable to connect with Node[%s][%s:%s].\n",nodeAddress[0],nodeAddress[1],nodeAddress[2])
	}
}

func (client *Client) setDbFilename(filename string){
	client.dbHandler.dbFileName=client.mainServerDir+"/"+filename
}

// only for test
func defaultClientObj() Client {
	client:=Client{}
	client.serverPORT="12349"
	client.serverAddress="127.0.0.1"
	client.address="127.0.0.1"
	client.PORT="54321"
	client.createWorkingDir()
	client.setDbFilename("client_DB.db")
	client.dbHandler.setupDb()
	return client
}

func main(){
	//fileHandler:=FileHandler{}
	//
	//fileHandler.FileName="a.jpg"
	//fmt.Println(fileHandler.hash())
	//
	//fileHandler.FileName="xa.png"
	//fmt.Println(fileHandler.hash())
	//
	//fileHandler.FileName="zzz.pdf"
	//fmt.Println(fileHandler.hash())
	//
	//fmt.Println()

	//---------------TEST------------//
	//client:=defaultClientObj()

	//client.sendFileToServer("/home/maroc/Videos/ds_test_files/a.jpg")
	//client.sendToBestLocationForFile("/home/maroc/Videos/ds_test_files/a.jpg")
	//client.getFileFromServer("a.jpg")
	//---------------------------//


	//files /home/maroc/Videos/d.jpg /home/maroc/Downloads/e.png  "/home/maroc/Downloads/master.pdf"
	// /home/maroc/Videos/ds_test_files/a.jpg

}