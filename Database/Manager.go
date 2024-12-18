package Database

import (
	"encoding/json"
	"fmt"
	"github.com/oracle/nosql-go-sdk/nosqldb"
	"github.com/oracle/nosql-go-sdk/nosqldb/auth/iam"
	"github.com/oracle/nosql-go-sdk/nosqldb/common"
	"github.com/oracle/nosql-go-sdk/nosqldb/types"
	"time"
)

type Database interface {
	Connect() *DataError
	Close() *DataError
	SelectTable(string) *DataError
	Get(string, string) (*DataError, *string)
	Put(interface{}) *DataError
	Delete(string, string) *DataError
	CleanupCache()
}

type OracleDB struct {
	Connected  bool
	ConfigPath string
	Region     common.Region
	Client     *nosqldb.Client
	Table      *string
	cache      *Cache[string, any]
}

func (db *OracleDB) Connect() *DataError {
	AuthProv, err := iam.NewSignatureProviderFromFile(db.ConfigPath, "", "", "")
	if err != nil {
		trace := NewTraceErr("Connection:", err.Error())
		return NewDataError(trace, "Connection")
	}

	cfg := nosqldb.Config{
		Region:                db.Region,
		AuthorizationProvider: AuthProv,
	}

	client, err := nosqldb.NewClient(cfg)
	if err != nil {
		trace := NewTraceErr("Client Creation:", err.Error())
		return NewDataError(trace, "Creation")
	}

	db.Client = client
	db.Connected = true
	return nil
}

func (db *OracleDB) Close() *DataError {
	if db.Connected == false {
		trace := NewTraceErr("Client Closing:", "No connection is open")
		return NewDataError(trace, "No Connection")
	}

	if db.Client == nil {
		trace := NewTraceErr("Client Closing:", "No client to close")
		return NewDataError(trace, "No Client")
	}

	err := db.Client.Close()
	if err != nil {
		trace := NewTraceErr("Client Closing:", err.Error())
		return NewDataError(trace, "Closing")
	}
	db.Connected = false
	return nil
}

func (db *OracleDB) SelectTable(name string) *DataError {
	db.Table = &name
	return nil
}

func (db *OracleDB) Get(Field string, Value string) (*DataError, *string) {
	callTrace := fmt.Sprintf("\n\tTable:\t%s\n\tGet:\t%s:\t%s", *db.Table, Field, Value)
	if !db.Connected {
		trace := NewTraceErr(callTrace, "Client not connected")
		return NewDataError(trace, "No Connection"), nil
	}

	if db.Table == nil {
		trace := NewTraceErr(callTrace, "Table is nil")
		return NewDataError(trace, "No Table Selected"), nil
	}

	v, ok := db.cache.Get(Value)
	if ok {
		fmt.Println("Cache Hit")
		jsonBytes, _ := json.Marshal(v)
		jsonString := string(jsonBytes)
		return nil, &jsonString
	}
	fmt.Println("Cache Miss")

	keyMap := &types.MapValue{}
	keyMap.Put(Field, Value)
	req := &nosqldb.GetRequest{
		TableName: *db.Table,
		Key:       keyMap,
	}
	res, err := db.Client.Get(req)
	if err != nil {
		trace := NewTraceErr(callTrace, err.Error())
		return NewDataError(trace, "Get Error"), nil
	}
	if !res.RowExists() {
		trace := NewTraceErr(callTrace, "No row exists")
		return NewDataError(trace, "No Row"), nil
	}
	json := res.ValueAsJSON()
	db.cache.Set(Value, json)
	fmt.Println("Cache Updated!")
	return nil, &json
}

func (db *OracleDB) Put(record interface{}) *DataError {
	callTrace := fmt.Sprintf("\n\tTable:\t%s\n\tPut:\t%+v", *db.Table, record)
	if !db.Connected {
		trace := NewTraceErr(callTrace, "Client not connected")
		return NewDataError(trace, "No Connection")
	}

	if db.Table == nil {
		trace := NewTraceErr(callTrace, "Table is nil")
		return NewDataError(trace, "No Table Selected")
	}

	if record == nil {
		trace := NewTraceErr(callTrace, "Record is empty")
		return NewDataError(trace, "No Record")
	}

	putRq := &nosqldb.PutRequest{
		TableName:   *db.Table,
		StructValue: record,
		ExactMatch:  true,
	}
	res, err := db.Client.Put(putRq)
	if err != nil {
		trace := NewTraceErr(callTrace, err.Error())
		return NewDataError(trace, "Put Error")
	}

	if !res.Success() {
		why := fmt.Sprintf("Put failed: %s", res.String())
		trace := NewTraceErr(callTrace, why)
		return NewDataError(trace, "Put Error")
	}
	return nil
}

func (db *OracleDB) Delete(key string, val string) *DataError {
	callTrace := fmt.Sprintf("\n\tTable:\t%s\n\tDelete:\t%s :\t%s", *db.Table, key, val)
	if !db.Connected {
		trace := NewTraceErr(callTrace, "Client not connected")
		return NewDataError(trace, "No Connection")
	}

	if db.Table == nil {
		trace := NewTraceErr(callTrace, "Table is nil")
		return NewDataError(trace, "No Table Selected")
	}

	deleteMap := &types.MapValue{}
	deleteMap.Put(key, val)
	deleteRq := &nosqldb.DeleteRequest{
		TableName: *db.Table,
		Key:       deleteMap,
	}

	res, err := db.Client.Delete(deleteRq)
	if err != nil {
		trace := NewTraceErr(callTrace, err.Error())
		return NewDataError(trace, "Delete Error")
	}

	if !res.Success {
		why := fmt.Sprintf("Delete failed: %s", res.String())
		trace := NewTraceErr(callTrace, why)
		return NewDataError(trace, "Delete Error")
	}
	db.cache.Delete(key)
	return nil
}

func (db *OracleDB) CleanupCache() {
	db.cache.Cleanup()
}

type TraceErr struct {
	message string
}

func (e *TraceErr) Error() string {
	return fmt.Sprintf("[DatabaseError] %s", e.message)
}

func NewTraceErr(where string, what string) *TraceErr {
	msg := fmt.Sprintf("\t%s \n\t%s", where, what)
	return &TraceErr{message: msg}
}

func NewDataError(err *TraceErr, Type string) *DataError {
	return &DataError{
		error: err,
		Type:  Type,
	}
}

type DataError struct {
	error
	Type string
}

func (e *DataError) ErrorType() string {
	return e.Type
}

func NewOracleConnection(filepath string, region common.Region, cacheTTL time.Duration) *OracleDB {
	return &OracleDB{
		Connected:  false,
		ConfigPath: filepath,
		Region:     region,
		Client:     nil,
		Table:      nil,
		cache:      NewCache[string, any](cacheTTL),
	}
}
