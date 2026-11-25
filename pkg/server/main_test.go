package server

import (
	"os"
	"testing"

	"github.com/SEG-UNIBE/artio-relay/pkg/storage/handlers"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
TestMain replaces the database used for request logging with a dry run connection,
so the tests do not need a running database (without one the log handler has a nil connection)
*/
func TestMain(m *testing.M) {
	dryRun, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=localhost"}), &gorm.Config{
		DryRun:                 true,
		SkipDefaultTransaction: true,
		DisableAutomaticPing:   true,
		Logger:                 logger.Discard,
	})
	if err != nil {
		panic(err)
	}
	handlers.LogHandlerObject = handlers.EventHandler{BaseHandler: &handlers.BaseHandler{Connection: dryRun}}

	os.Exit(m.Run())
}
