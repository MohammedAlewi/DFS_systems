package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)


type FileEntry struct{
	dbFileName string
	entryDate  string
	dbConn     *sql.DB
	nodeAddress []string
}

func (fileEntry *FileEntry) setupDb() {
	db, err := sql.Open("sqlite3", fileEntry.dbFileName)
	fileEntry.checkError(err)
	db.Exec(`CREATE TABLE IF NOT EXISTS  known_file_locations (node_id TEXT, file_name TEXT,node_port TEXT, node_ip TEXT, entry_date datetime)`)

	fileEntry.dbConn =db
	time.Sleep(time.Second)
}


func (fileEntry *FileEntry) insertRoutingAddress(nodeID,nodeAddress,nodePort,fileName string){
	sqlAddItem := `INSERT OR REPLACE INTO known_file_locations(node_id,file_name,node_port,node_ip,entry_date) values(?, ?,?,?, CURRENT_TIMESTAMP)`
	stmt, err := fileEntry.dbConn.Prepare(sqlAddItem)
	fileEntry.checkError(err)
	_, err2 := stmt.Exec(nodeID,fileName,nodePort,nodeAddress)
	fileEntry.checkError(err2)
	defer stmt.Close()
}

func (fileEntry *FileEntry) getFileEntries(fileName string) bool {
	nodeAddress:=[]string{"",""}
	sqlStatement := `SELECT node_ip, node_port FROM known_file_locations WHERE file_name=$1;`
	row:=fileEntry.dbConn.QueryRow(sqlStatement, fileName)
	err:=row.Scan(nodeAddress[0],nodeAddress[1])
	fileEntry.nodeAddress=nodeAddress
	if err!=nil { return false}
	return true
}

func (fileEntry *FileEntry) getAllRoutingID() [][]string {
	ids:=[][]string{}
	statement, err := fileEntry.dbConn.Prepare("SELECT node_id,node_ip,node_port FROM known_file_locations;")
	fileEntry.checkError(err)
	rows, err := statement.Query()
	defer rows.Close()
	for rows.Next() {
		nodeAdd:=[]string{"","",""}
		rows.Scan(&nodeAdd[0],&nodeAdd[1],&nodeAdd[2])
		ids=append(ids, nodeAdd)
	}
	return ids
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
