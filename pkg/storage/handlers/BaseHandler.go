package handlers

import (
	"artio-relay/pkg/config"
	"artio-relay/pkg/storage/models"
	"bytes"
	"encoding/json"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type IHandler interface {
	GetAll() (any, error)
	Get(id uint) (any, error)
	DeleteAll() (*gorm.DB, error)
	Delete(id uint) (*gorm.DB, error)
	Create(values []byte) (any, error)
}

/*
BaseHandler Object defintion with parameters
*/
type BaseHandler struct {
	IHandler
	objects    any
	Connection *gorm.DB
	tableName  string
}

/*
Get fetch a given Event by id
*/
func (b *BaseHandler) Get(id uint) (any, error) {
	_ = b.Connection.First(b.objects, id)
	return b.objects, nil
}

/*
GetAll Fetch all the events from the database
*/
func (b *BaseHandler) GetAll() (any, error) {
	_ = b.Connection.Find(b.objects)
	return b.objects, nil
}

/*
DeleteAll Function to delete all entries from a single table
*/
func (b *BaseHandler) DeleteAll() (*gorm.DB, error) {
	results := b.Connection.Delete(b.objects)
	return results, nil
}

/*
Delete Function to delete a single instance out of the database
*/
func (b *BaseHandler) Delete(id uint) (*gorm.DB, error) {
	results := b.Connection.Delete(b.objects, id)
	return results, nil
}

/*
Create a new object in the storage
*/
func (b *BaseHandler) Create(values []byte) (any, error) {
	object := models.Event{} // TODO: handle this case properly
	err := json.NewDecoder(bytes.NewBuffer(values)).Decode(&object)
	if err != nil {
		return nil, err
	}
	b.Connection.Table("events") // TODO: Fix this stupid thing that will no work
	results := b.Connection.Create(&object)
	return results, nil
}

/*
NewBaseHandler Function to create a new BaseHandler with all the necessary parameters
*/
func NewBaseHandler[T any](objects []T) BaseHandler {
	dsn := config.Config.GetDatabaseConnectionString()
	dbConn, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return BaseHandler{Connection: dbConn, objects: objects}
}
