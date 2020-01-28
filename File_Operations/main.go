package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)


type FileEntry struct{
	hashValue  int
	fileName   string
	dbFileName string
	entryDate  string
	dbConn     *sql.DB
}

func (fileEntry *FileEntry) setDbFilename(fileName string){
	fileEntry.dbFileName =fileName
}

func (fileEntry *FileEntry) setHashValue(hash int){
	fileEntry.hashValue =hash
}

func (fileEntry *FileEntry) setFileName(fileName string){
	fileEntry.fileName =fileName
}


func (fileEntry *FileEntry) setupDb() {
	os.Remove(fileEntry.dbFileName)
	db, err := sql.Open("sqlite3", fileEntry.dbFileName)
	fileEntry.checkError(err)
	db.Exec(`create table saved_files (hash_value integer, file_name TEXT, entry_date datetime)`)
	fileEntry.dbConn =db
}

func (fileEntry *FileEntry) openDb(){
	db, err := sql.Open("sqlite3", fileEntry.dbFileName)
	fileEntry.checkError(err)
	fileEntry.dbConn =db
}


func (fileEntry *FileEntry) insertFileEntry(){
	sqlAdditem := `INSERT OR REPLACE INTO saved_files(hash_value,file_name,entry_date) values(?, ?, CURRENT_TIMESTAMP)`
	stmt, err := fileEntry.dbConn.Prepare(sqlAdditem)
	fileEntry.checkError(err)
	_, err2 := stmt.Exec(fileEntry.hashValue, fileEntry.fileName)
	fileEntry.checkError(err2)
	defer stmt.Close()
}


func (fileEntry *FileEntry) getFileEntries(hashValue int) bool {
	sqlStatement := `SELECT hash_value, file_name, entry_date FROM saved_files WHERE hash_value=$1;`
	row:=fileEntry.dbConn.QueryRow(sqlStatement,hashValue)
	err:=row.Scan(&fileEntry.hashValue,&fileEntry.fileName,&fileEntry.entryDate)
	fileEntry.checkError(err)
	return true
}

func (fileEntry *FileEntry) closeDb(){
	fileEntry.dbConn.Close()
}

func (fileEntry *FileEntry) checkError(err error){
	if err!=nil{
		panic(err)
	}
}

func main() {
	//fileEntry:=FileEntry{}
	//fileEntry.setDbFilename("FileEntries.db")
	//fileEntry.setHashValue(2323)
	//fileEntry.setFileName("ababa.jpg")
	//
	//fileEntry.setupDb()
	//fileEntry.insertFileEntry()
	//fileEntry.closeDb()

	//fileEntry.setDbFilename("FileEntries.db")
	//fileEntry.openDb()
	//fileEntry.getFileEntries(2323)
	//
	//fmt.Println(fileEntry.entryDate,fileEntry.hashValue,fileEntry.fileName)


}


