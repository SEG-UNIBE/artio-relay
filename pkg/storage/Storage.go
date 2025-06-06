package storage

import (
	"artio-relay/pkg/config"
	"artio-relay/pkg/storage/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/*
IStorage Interface specification for the Storage
*/
type IStorage interface {
	init() error
}

/*
Storage object definition of the storage with params
*/
type Storage struct {
	Connection *gorm.DB
}

/*
Init Initialize the storage and set up the model
*/
func (s *Storage) Init() error {
	connection, err := gorm.Open(postgres.Open(config.Config.GetDatabaseConnectionString()), &gorm.Config{})
	if err != nil {
		return err
	}
	s.Connection = connection
	setupErr := s.setupModel()
	if setupErr != nil {
		return setupErr
	}
	return nil
}

/*
setupModel initializes the newest model in the database
*/
func (s *Storage) setupModel() error {
	migrateErr := s.Connection.AutoMigrate(&models.Event{}, &models.Log{})
	if migrateErr != nil {
		return migrateErr
	}
	return nil
}
