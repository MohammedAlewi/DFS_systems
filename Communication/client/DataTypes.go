package main

type Message struct {
	SenderIP string
	SenderPort string
	FileName   string
	NodeID	uint32
	Data []byte
	SIZE int
	COMMAND string
}


//enum

const(
	SAVE_FILE  = "SAVE_FILE"
	GET_FILE = "GET_FILE"
	FILE_NOT_FOUND = "FILE_NOT_FOUND"
	FILE_FOUND = "FILE_FOUND"
)