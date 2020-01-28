package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)


type FileEntry struct{
	hashValue  string
	fileName   string
	dbFileName string
	entryDate  string
	dbConn     *sql.DB
}

func (fileEntry *FileEntry) setupDb() {
	db, err := sql.Open("sqlite3", fileEntry.dbFileName)
	fileEntry.checkError(err)
	//db.Exec(`create table saved_files (hash_value TEXT, file_name TEXT, entry_date datetime)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS  saved_files (hash_value TEXT, file_name TEXT, entry_date datetime)`)

	//db.Exec(`create table routing_table (node_id TEXT, node_address TEXT, entry_date datetime)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS  routing_table (node_id TEXT, node_address TEXT,node_port TEXT, entry_date datetime)`)
	fileEntry.dbConn =db
	time.Sleep(time.Second)
}


func (fileEntry *FileEntry) setRecordState(hashValue string ,fileName string){
	fileEntry.hashValue = hashValue
	fileEntry.fileName=fileName
}

func (fileEntry *FileEntry) insertFileEntry(){
	sqlAddItem := `INSERT OR REPLACE INTO saved_files(hash_value,file_name,entry_date) values(?, ?, CURRENT_TIMESTAMP)`
	stmt, err := fileEntry.dbConn.Prepare(sqlAddItem)
	fileEntry.checkError(err)
	_, err2 := stmt.Exec(fileEntry.hashValue, fileEntry.fileName)
	fileEntry.checkError(err2)
	defer stmt.Close()
}

func (fileEntry *FileEntry) insertRoutingAddress(nodeID,nodeAddress,nodePort string){
	sqlAddItem := `INSERT OR REPLACE INTO routing_table(node_id,node_address,node_port,entry_date) values(?, ?,?, CURRENT_TIMESTAMP)`
	stmt, err := fileEntry.dbConn.Prepare(sqlAddItem)
	fileEntry.checkError(err)
	_, err2 := stmt.Exec(nodeID, nodeAddress,nodePort)
	fileEntry.checkError(err2)
	defer stmt.Close()
}

func (fileEntry *FileEntry) getFileEntries(hashValue string) bool {
	sqlStatement := `SELECT hash_value, file_name, entry_date FROM saved_files WHERE hash_value=$1;`
	row:=fileEntry.dbConn.QueryRow(sqlStatement,hashValue)
	err:=row.Scan(&fileEntry.hashValue,&fileEntry.fileName,&fileEntry.entryDate)
	if err!=nil { return false}
	return true
}

func (fileEntry *FileEntry) getAllRoutingID() []string {
	ids:=[]string{}
	statement, err := fileEntry.dbConn.Prepare("SELECT node_id FROM routing_table;")
	fileEntry.checkError(err)
	rows, err := statement.Query()
	defer rows.Close()
	var nodeId string
	for rows.Next() {
		rows.Scan(&nodeId)
		ids=append(ids, nodeId)
	}
	return ids
}

func (fileEntry *FileEntry) getAllFileEntries() []string {
	ids:=[]string{}
	statement, err := fileEntry.dbConn.Prepare("SELECT hash_value FROM saved_files;")
	fileEntry.checkError(err)
	rows, err := statement.Query()
	defer rows.Close()
	var fileId string
	for rows.Next() {
		rows.Scan(&fileId)
		ids=append(ids, fileId)
	}
	return ids
}

func  (fileEntry *FileEntry) removeFileEntry(hashValue string) bool{
	sqlDel, _ := fileEntry.dbConn.Begin()
	stmt, _ := sqlDel.Prepare("DELETE FROM saved_files WHERE hash_value=?;")
	defer stmt.Close()
	_, err := stmt.Exec(hashValue)
	if err!=nil { return false}
	sqlDel.Commit()
	return true
}

func (fileEntry *FileEntry) getRoutingAddress(nodeAddress string) []string {
	address,port:="",""
	sqlStatement := `SELECT node_address,node_port FROM routing_table WHERE node_id=$1;`
	row:=fileEntry.dbConn.QueryRow(sqlStatement,nodeAddress)
	err:=row.Scan(&address,&port)
	if err!=nil { return []string{}}
	return []string{address,port}
}

func (fileEntry *FileEntry) closeDb(){
	fileEntry.dbConn.Close()
}

func (fileEntry *FileEntry) checkError(err error){
	if err!=nil{
		panic(err)
	}
}
//  Just to print what is inside those tables......
func (fileEntry *FileEntry) printAllFileEntries(){
	statement, err := fileEntry.dbConn.Prepare("SELECT hash_value, file_name, entry_date FROM saved_files;")
	fileEntry.checkError(err)

	rows, err := statement.Query()
	defer rows.Close()

	fmt.Println("File Entries:")
	fmt.Println("-------------------------------------------------")

	for rows.Next() {
		var hashValue int
		var fileName string
		var entryDate string
		rows.Scan(&hashValue, &fileName,&entryDate)
		fmt.Printf("[%d]:  %v : %v\n", hashValue,fileName, entryDate)
	}
	fmt.Println("-------------------------------------------------")
}


func (fileEntry *FileEntry) printAllRouteEntries(){
	statement, err := fileEntry.dbConn.Prepare("SELECT node_id,node_address,entry_date FROM routing_table;")
	fileEntry.checkError(err)

	rows, err := statement.Query()
	defer rows.Close()

	fmt.Println("Route Entries:")
	fmt.Println("-------------------------------------------------")

	for rows.Next() {
		var nodeId int
		var nodeAddress,entryDate string
		rows.Scan(&nodeId, &nodeAddress,&entryDate)
		fmt.Printf("[%d]:  %v : %v\n", nodeId, nodeAddress, entryDate)
	}
	fmt.Println("-------------------------------------------------")
}
