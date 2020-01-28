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

type SERVER struct {
	serverID    uint32
	portNo      string
	fileHandler FileHandler
	dbHandler  FileEntry
	listener net.Listener
	wg sync.WaitGroup
	dbFileName string
	connection net.Conn
	mainServerDir string
	// RECOVERY
	rebroadcastedMessages []uint32
}

func (server *SERVER) createListenerConnection() { //net.Listener
	fmt.Printf("[NODE %d]: listening connection from a client or remote node\n",server.serverID)

	listener, err := net.Listen("tcp", ":"+server.portNo)
	server.listener=listener
	server.checkError(err)
}

func (server *SERVER) createConnection(clientAdd string,port string) bool {
	server.wg.Add(1)
	conn, err:= net.Dial("tcp",clientAdd+":"+port)
	if err!=nil{
		fmt.Printf("[NODE %d]: unable to connected with a remote node\n",server.serverID)
		return false
	}
	server.connection=conn

	fmt.Printf("[NODE %d]: connected with a client node\n",server.serverID)
	return true
}

func (server *SERVER) getClientRequest(hashValue string) Message {
	server.dbHandler.dbFileName=server.dbFileName
	server.dbHandler.setupDb()
	result:=server.dbHandler.getFileEntries(hashValue)
	if result{
		fmt.Printf("[NODE %d]: File Found [%s] %s\n",server.serverID,server.dbHandler.hashValue,server.dbHandler.fileName)
		server.fileHandler.FileName=server.mainServerDir+"/Received_Files/"+hashValue+"__"+server.dbHandler.fileName
		message:=server.fileHandler.readFile()
		message.FileName=server.dbHandler.fileName
		message.NodeID=server.serverID
		return message
	}else{
		fmt.Printf("[NODE %d]: NOT FOUND\n",server.serverID)
		return Message{}
	}
}

func (server *SERVER) sendMessageToClientNode(messageObj Message){
	defer server.wg.Done()
	var l sync.Mutex
	l.Lock()
	binBuf := new(bytes.Buffer)
	goBufferObj := gob.NewEncoder(binBuf)
	goBufferObj.Encode(messageObj)

	_,err2:=server.connection.Write(binBuf.Bytes())
	if err2!=nil{
		fmt.Printf("[NODE %d]: Message not sent to the client.\n",server.serverID)
	}else {
		server.connection.Close()
		fmt.Printf("[NODE %d]: Message sent to the client.\n",server.serverID)
	}
	l.Unlock()
}

func (server *SERVER) acceptConnection(){
	for true==true{
		server.wg.Add(1)
		conn,err := server.listener.Accept()

		server.checkError(err)
		go server.handleConnection(conn)

	}
	server.wg.Wait()
}

func (server *SERVER) receiveMessageFromClient(message Message){
	var l sync.Mutex
	l.Lock()
	// file saved at the server
	server.fileHandler.setFileHandler(message)
	server.fileHandler.writeFile(server.mainServerDir+"/Received_Files")

	// file hash and access  path saved at the database
	server.dbHandler.dbFileName=server.dbFileName
	server.dbHandler.setupDb()
	server.dbHandler.setRecordState(server.fileHandler.hash(),server.fileHandler.FileName)
	server.dbHandler.insertFileEntry()

	fmt.Printf("[NODE %d]: file %s have been accepted and saved in local working dir and database\n",server.serverID,message.FileName)
	l.Unlock()
}


func (server *SERVER) handleConnection(conn net.Conn){
	defer server.wg.Done()

	dec := gob.NewDecoder(conn)
	messageObject := &Message{}
	dec.Decode(messageObject)
	conn.Close()

	switch messageObject.COMMAND {
		case  SAVE_FILE:{
			server.chooseBestNodeToSaveFile(*messageObject)
		}
		case  GET_FILE:{
			server.chooseBestNodeToPropagateRequest(*messageObject,conn)
		}
		case FILE_MIGRATION_REQUEST:{
			server.checkMigrationFiles(*messageObject)
			server.rebroadcastMessages(*messageObject)
		}
		case FILE_MIGRATION_RESPONSE:{
			server.receiveMessageFromClient(*messageObject)
		}
	}
}

func (server *SERVER) closeListener(){
	server.listener.Close()
}

func (server *SERVER) checkError(err error){
	if err!=nil{
		panic(err)
	}
}

// routing's
func (server *SERVER) getBestNode(fileHashValue string) map[int]string {
	nodeIds:=server.dbHandler.getAllRoutingID();
	weights:=make(map[int]string)

	fileBinHash,_:=strconv.ParseUint(fileHashValue, 10, 32)
	for _,routeID := range nodeIds{
		intId,_:=strconv.ParseUint(routeID, 10, 32)
		currentWeight:=intId ^ fileBinHash
		weights[int(currentWeight)]= routeID
	}
	return weights
}

func (server *SERVER) chooseBestNodeToSaveFile(messageObject Message){
	server.fileHandler.FileName=messageObject.FileName
	fileHashValue:=server.fileHandler.hash()

	nodes:=server.getBestNode(fileHashValue)
	keys:=[]int{}
	for  key := range nodes  {
		keys=append(keys, key)
	}
	sort.Ints(keys)
	// calculating self weight.....
	fileBinHash,_:=strconv.ParseUint(fileHashValue, 10, 32)
	selfWeight := fileBinHash^uint64(server.serverID)
	server.propagateFileSaveRequest(messageObject,selfWeight,keys,nodes)
}

func (server *SERVER) propagateFileSaveRequest(messageObject Message,selfWeight uint64,keys []int,nodes map[int]string){
	for key:=0; key < len(keys);key++{
		if len(keys)==0 || int(selfWeight)<= keys[key]{
			fmt.Printf("[NODE %d]: current Node is the best node to save the file \n",server.serverID)

			server.receiveMessageFromClient(messageObject)
			break
		}else {
			fmt.Printf("[NODE %d]: current Node is NOT the best node to save the file forwarding the request \n", server.serverID)

			nodeAddress := server.dbHandler.getRoutingAddress(nodes[keys[key]])
			result := server.createConnection(nodeAddress[0], nodeAddress[1])
			if result {
				server.sendMessageToClientNode(messageObject)
				break
			}
			fmt.Printf("[NODE %d]: Node [%s] is currently not responding trying the next best server.... \n",
				server.serverID, nodes[keys[key]])
			if key+1==len(keys){
				fmt.Printf("[NODE %d]: no choice found so current Node is the best node to save the file \n",server.serverID)
				server.receiveMessageFromClient(messageObject)
			}
		}
	}
}

func (server *SERVER) chooseBestNodeToPropagateRequest(messageObject Message,conn net.Conn){
	server.fileHandler.FileName=messageObject.FileName
	fileHashValue:=server.fileHandler.hash()

	nodes:=server.getBestNode(fileHashValue)

	keys:=[]int{}
	for  key := range nodes  {
		keys=append(keys, key)
	}
	sort.Ints(keys)

	fileBinHash,_:=strconv.ParseUint(fileHashValue, 10, 32)
	selfWeight := fileBinHash^uint64(server.serverID)
	server.propagateFileQueryRequest(messageObject,conn,selfWeight,fileHashValue,keys,nodes)
}

func (server *SERVER) propagateFileQueryRequest(messageObject Message,conn net.Conn,selfWeight uint64,fileHashValue string,keys []int,nodes map[int]string,){
	for key,_:= range keys{
		if len(keys)==0 || int(selfWeight)<= keys[key]{
			fmt.Printf("[NODE %d]: current Node is the best node to get the file from \n",server.serverID)
			server.getFileFromCurrentNode(fileHashValue,conn,messageObject)
			break
		}else{
			fmt.Printf("[NODE %d]: current Node is NOT the best node to get the file from forwarding the request... \n",server.serverID)
			nodeAddress:= server.dbHandler.getRoutingAddress(nodes[keys[key]])
			result:=server.createConnection(nodeAddress[0],nodeAddress[1])
			if result{server.sendMessageToClientNode(messageObject)
				break
			}
			fmt.Printf("[NODE %d]: Node [%s] is currently not responding trying the next best server.... \n",
				server.serverID, nodes[keys[key]])
			if key+1==len(keys){
				fmt.Printf("[NODE %d]: no choice found so checking current Node for the file \n",server.serverID)
				server.getFileFromCurrentNode(fileHashValue,conn,messageObject)
			}
		}
	}
}

func  (server *SERVER) getFileFromCurrentNode( fileHashValue string,conn net.Conn,messageObject Message){
	msg:=server.getClientRequest(fileHashValue)
	msg.SenderIP=strings.Split(conn.LocalAddr().String(),":")[0]
	msg.SenderPort=server.portNo
	if len(msg.FileName)!=0{
		server.createConnection(messageObject.SenderIP,messageObject.SenderPort)
		msg.COMMAND=FILE_FOUND
		server.sendMessageToClientNode(msg)
	}else{
		server.createConnection(messageObject.SenderIP,messageObject.SenderPort)
		msg.COMMAND=FILE_NOT_FOUND
		msg.FileName=messageObject.FileName
		server.sendMessageToClientNode(msg)
	}
}

// ------------------------ failure recovery -----------------------------  //
func (server *SERVER) broadCastMigrationRequest(currentIp string){
	routIds:=server.dbHandler.getAllRoutingID()
	message:=Message{SenderPort:server.portNo,
		SenderIP:currentIp,NodeID:server.serverID,COMMAND:FILE_MIGRATION_REQUEST}
	server.rebroadcastedMessages=append(server.rebroadcastedMessages,server.serverID)
	for _,routeId := range routIds{
		nodeAddress:=server.dbHandler.getRoutingAddress(routeId)
		result:=server.createConnection(nodeAddress[0],nodeAddress[1])
		if result{server.sendMessageToClientNode(message)
			fmt.Printf("[NODE %d]: Broadcasting File Migration request to node[%s]... \n",server.serverID,routeId)
		}
	}
}

func (server *SERVER) checkMigrationFiles(messageObj Message){
	var l sync.Mutex
	l.Lock()

	fileIds:=server.dbHandler.getAllFileEntries();
	for _,nodeId:= range server.rebroadcastedMessages{
		fmt.Println("explored",nodeId,messageObj.NodeID)
		if nodeId==messageObj.NodeID{return}
	}
	for _,fileID := range fileIds{
		intId,_:=strconv.ParseUint(fileID, 10, 32)
		requestWeight:=intId ^ uint64(messageObj.NodeID)
		currentWeight:=intId ^ uint64(server.serverID)

		if requestWeight< currentWeight{
			result := server.createConnection(messageObj.SenderIP, messageObj.SenderPort)
			if result {
				message:=server.getClientRequest(fileID)
				message.COMMAND=FILE_MIGRATION_RESPONSE
				server.sendMessageToClientNode(message)
				server.dbHandler.removeFileEntry(fileID)
				server.fileHandler.removeFileFromDir(server.mainServerDir+"/Received_Files/"+fileID+"__"+message.FileName)
				fmt.Printf("[NODE %d]: Migrating File [%s] %s to the Node[%d]... \n",
					server.serverID,fileID,message.FileName,messageObj.NodeID)
			}
		}
	}
	l.Unlock()
}

func (server *SERVER) rebroadcastMessages(message Message){
	for _,nodeId:= range server.rebroadcastedMessages{
		if nodeId==message.NodeID{return}
	}
	server.rebroadcastedMessages=append(server.rebroadcastedMessages,message.NodeID)
	routIds:=server.dbHandler.getAllRoutingID()
	for _,routeId := range routIds{
		nodeAddress:=server.dbHandler.getRoutingAddress(routeId)
		result:=server.createConnection(nodeAddress[0],nodeAddress[1])
		if result{server.sendMessageToClientNode(message)
			fmt.Printf("[NODE %d]: rebroadcasting messages to node [%s] \n",server.serverID,routeId)
		}
	}
}

//--------------------- setup server environment ------------------------- //
func (server *SERVER) updateRoutingtable(routes [][]string){
	for _, route := range routes {
		server.dbHandler.insertRoutingAddress(route[0],route[1],route[2])
	}
}

func (server *SERVER) setDbFilename(filename string){
	server.dbFileName=server.mainServerDir+"/"+filename
	server.dbHandler.dbFileName=server.mainServerDir+"/"+filename
}

func (server *SERVER) createWorkingDir() {
	dir:= strconv.FormatUint(uint64(server.serverID),10)+"_SERVER_FILES/Received_Files"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
	server.mainServerDir=strconv.FormatUint(uint64(server.serverID),10)+"_SERVER_FILES"
}


// interactions
func setupServer(port string,serverID uint32,routeTableInfo [][]string) SERVER {
	server:=SERVER{}
	server.portNo=port
	server.serverID=serverID
	server.createWorkingDir()
	server.setDbFilename(fmt.Sprint(serverID)+"_serverData.db")
	server.dbHandler.setupDb()
	server.updateRoutingtable(routeTableInfo)
	return  server
}

// .................. create server object .................... //
func creatServerInstance(port string, serverid uint32,routes [][]string){
	defer wgOut.Done()  // with synchronization setup the caller function
	server:=setupServer(port,serverid,routes)
	server.createListenerConnection()
	server.acceptConnection()
	server.closeListener()
}


func creatFailedServerInstance(port string, serverIp string, serverid uint32,routes [][]string){
	defer wgOut.Done()  // with synchronization setup the caller function
	server:=setupServer(port,serverid,routes)
	server.createListenerConnection()
	server.broadCastMigrationRequest(serverIp)
	server.acceptConnection()
	server.closeListener()
}






