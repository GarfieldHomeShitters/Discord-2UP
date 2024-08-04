package Database

import (
	"fmt"
	"github.com/oracle/nosql-go-sdk/nosqldb"
	"github.com/oracle/nosql-go-sdk/nosqldb/auth/iam"
	"github.com/oracle/nosql-go-sdk/nosqldb/common"
	"github.com/oracle/nosql-go-sdk/nosqldb/types"
)

type Database interface {
	Connect() error
	Close() error
	SelectTable(string) error
	Get(string, string) (error, *string)
	Put(interface{}) error
	Delete(string, string) error
}

type OracleDB struct {
	Connected  bool
	ConfigPath string
	Region     common.Region
	Client     *nosqldb.Client
	Table      *string
}

func (db *OracleDB) Connect() error {
	AuthProv, err := iam.NewSignatureProviderFromFile(db.ConfigPath, "", "", "")
	if err != nil {
		return NewDatabaseError("Connection:", err.Error())
	}

	cfg := nosqldb.Config{
		Region:                db.Region,
		AuthorizationProvider: AuthProv,
	}

	client, err := nosqldb.NewClient(cfg)
	if err != nil {
		return NewDatabaseError("Client Creation:", err.Error())
	}

	db.Client = client
	return nil
}

func (db *OracleDB) Close() error {
	if db.Connected == false {
		return NewDatabaseError("Client Closing:", "No connection is open")
	}

	if db.Client == nil {
		return NewDatabaseError("Client Closing:", "No client to close")
	}

	err := db.Client.Close()
	if err != nil {
		return NewDatabaseError("Client Closing:", err.Error())
	}
	db.Connected = false

	return nil
}

func (db *OracleDB) SelectTable(name string) error {
	db.Table = &name
	return nil
}

func (db *OracleDB) Get(Field string, Value string) (error, *string) {
	callTrace := fmt.Sprintf("\n\tTable:\t%s\n\tGet:\t%s:\t%s", *db.Table, Field, Value)
	if !db.Connected {
		return NewDatabaseError(callTrace, "Client not connected"), nil
	}

	if db.Table == nil {
		return NewDatabaseError(callTrace, "Table is nil"), nil
	}

	keyMap := &types.MapValue{}
	keyMap.Put(Field, Value)
	req := &nosqldb.GetRequest{
		TableName: *db.Table,
		Key:       keyMap,
	}
	res, err := db.Client.Get(req)
	if err != nil {
		return NewDatabaseError(callTrace, err.Error()), nil
	}
	if !res.RowExists() {
		return NewDatabaseError(callTrace, "No row exists"), nil
	}
	json := res.ValueAsJSON()
	return nil, &json
}

func (db *OracleDB) Put(record interface{}) error {
	callTrace := fmt.Sprintf("\n\tTable:\t%s\n\tPut:\t%+v", *db.Table, record)
	if !db.Connected {
		return NewDatabaseError(callTrace, "Client not connected")
	}

	if db.Table == nil {
		return NewDatabaseError(callTrace, "Table is nil")
	}

	if record == nil {
		return NewDatabaseError(callTrace, "Record is empty")
	}

	putRq := &nosqldb.PutRequest{
		TableName:   *db.Table,
		StructValue: record,
		ExactMatch:  true,
	}
	res, err := db.Client.Put(putRq)
	if err != nil {
		return NewDatabaseError(callTrace, err.Error())
	}

	if !res.Success() {
		why := fmt.Sprintf("Put failed: %s", res.String())
		return NewDatabaseError(callTrace, why)
	}

	return nil

}

func (db *OracleDB) Delete(key string, val string) error {
	callTrace := fmt.Sprintf("\n\tTable:\t%s\n\tDelete:\t%s :\t%s", *db.Table, key, val)
	if !db.Connected {
		return NewDatabaseError(callTrace, "Client not connected")
	}
	if db.Table == nil {
		return NewDatabaseError(callTrace, "Table is nil")
	}

	deleteMap := &types.MapValue{}
	deleteMap.Put(key, val)
	deleteRq := &nosqldb.DeleteRequest{
		TableName: *db.Table,
		Key:       deleteMap,
	}

	res, err := db.Client.Delete(deleteRq)
	if err != nil {
		return NewDatabaseError(callTrace, err.Error())
	}

	if !res.Success {
		why := fmt.Sprintf("Delete failed: %s", res.String())
		return NewDatabaseError(callTrace, why)
	}

	return nil
}

type DbError struct {
	message string
}

func (e *DbError) Error() string {
	return fmt.Sprintf("[DatabaseError] %s", e.message)
}

func NewDatabaseError(where string, what string) *DbError {
	msg := fmt.Sprintf("\t%s \n\t%s", where, what)
	return &DbError{message: msg}
}

func NewOracleConnection(filepath string, region common.Region) *OracleDB {
	return &OracleDB{
		Connected:  false,
		ConfigPath: filepath,
		Region:     region,
		Client:     nil,
		Table:      nil,
	}
}
