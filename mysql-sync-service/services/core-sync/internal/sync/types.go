package sync

import (
	"fmt"
)

type EventType string

const (
	Insert EventType = "INSERT"
	Update EventType = "UPDATE"
	Delete EventType = "DELETE"
)

type BinlogEvent struct {
	Type      EventType
	Schema    string
	Table     string
	Rows      [][]interface{} // For Insert/Delete, or Update (old/new pairs)
	Timestamp uint32
	BinlogFile string
	BinlogPos  uint32
}

func (e BinlogEvent) String() string {
	return fmt.Sprintf("[%s] %s.%s (%d rows)", e.Type, e.Schema, e.Table, len(e.Rows))
}
