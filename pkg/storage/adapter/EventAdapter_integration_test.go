package adapter

import (
	"fmt"
	"testing"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/SEG-UNIBE/artio-relay/pkg/storage/handlers"
	"github.com/SEG-UNIBE/artio-relay/pkg/storage/models"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
useIntegrationDB points the storage handlers to the database from .env.tests, the test is skipped when it is not reachable
*/
func useIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	conf, err := (&config.Configuration{}).Init("../../../.env.tests")
	if err != nil {
		t.Skipf("skipping integration test, failed to load .env.tests: %v", err)
	}
	db, err := gorm.Open(postgres.Open(conf.GetDatabaseConnectionString()), &gorm.Config{Logger: logger.Discard})
	if err != nil || db == nil {
		t.Skipf("skipping integration test, database not reachable: %v", err)
	}
	// packages are tested in parallel, the advisory lock keeps them from migrating a fresh database at the same time
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(4242)").Error; err != nil {
			return err
		}
		return tx.AutoMigrate(&models.Event{}, &models.Log{})
	})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	originalEvents, originalLogs, originalConfig := handlers.EventHandlerObject, handlers.LogHandlerObject, *config.Config
	handlers.EventHandlerObject = handlers.EventHandler{BaseHandler: &handlers.BaseHandler{Connection: db}}
	handlers.LogHandlerObject = handlers.EventHandler{BaseHandler: &handlers.BaseHandler{Connection: db}}
	config.Config.RelayMaxMessageCount = 100
	t.Cleanup(func() {
		handlers.EventHandlerObject, handlers.LogHandlerObject, *config.Config = originalEvents, originalLogs, originalConfig
	})
	return db
}

/*
newPubkey returns a random pubkey whose events are removed after the test
*/
func newPubkey(t *testing.T, db *gorm.DB) string {
	t.Helper()
	pubkey := nostr.GeneratePrivateKey()
	t.Cleanup(func() {
		db.Unscoped().Where("pubkey = ?", pubkey).Delete(&models.Event{})
	})
	return pubkey
}

/*
createEvent stores a new event through the adapter
*/
func createEvent(t *testing.T, a *EventAdapter, pubkey string, kind int, created int64, tags nostr.Tags) nostr.Event {
	t.Helper()
	event := nostr.Event{ID: nostr.GeneratePrivateKey(), PubKey: pubkey, CreatedAt: nostr.Timestamp(created), Kind: kind, Tags: tags, Content: "content", Sig: "sig"}
	result, _ := a.Create(event)
	if err := result.(*gorm.DB).Error; err != nil {
		t.Fatalf("failed to create event: %v", err)
	}
	return event
}

/*
deletionRequest builds a kind 5 event with the given tags
*/
func deletionRequest(pubkey string, created int64, tags ...nostr.Tag) nostr.Event {
	return nostr.Event{ID: nostr.GeneratePrivateKey(), PubKey: pubkey, CreatedAt: nostr.Timestamp(created), Kind: 5, Tags: tags}
}

/*
remainingIDs returns the ids of all stored events of a pubkey
*/
func remainingIDs(t *testing.T, a *EventAdapter, pubkey string) map[string]bool {
	t.Helper()
	events, err := a.Get(nostr.Filter{Authors: []string{pubkey}, Limit: 100})
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}
	ids := make(map[string]bool)
	for _, event := range events {
		ids[event.ID] = true
	}
	return ids
}

/*
TestCreate_NilTags tests that events without tags are stored with an empty tag list
*/
func TestCreate_NilTags(t *testing.T) {
	db := useIntegrationDB(t)
	a := &EventAdapter{}
	pk := newPubkey(t, db)

	event := createEvent(t, a, pk, 1, 100, nil)

	events, err := a.Get(nostr.Filter{IDs: []string{event.ID}, Limit: 1})
	if err != nil || len(events) != 1 {
		t.Fatalf("expected the stored event but got %v (%v)", events, err)
	}
	if events[0].Tags == nil || len(events[0].Tags) != 0 {
		t.Fatalf("expected an empty tag list but got %v", events[0].Tags)
	}
}

/*
TestGet_ClampsLimit tests that the requested limit is capped at the configured maximum
*/
func TestGet_ClampsLimit(t *testing.T) {
	db := useIntegrationDB(t)
	config.Config.RelayMaxMessageCount = 2
	a := &EventAdapter{}
	pk := newPubkey(t, db)
	for i := int64(0); i < 4; i++ {
		createEvent(t, a, pk, 1, 100+i, nil)
	}

	events, err := a.Get(nostr.Filter{Authors: []string{pk}, Limit: 10})
	if err != nil || len(events) != 2 {
		t.Fatalf("expected 2 events but got %v (%v)", len(events), err)
	}
}

/*
TestDelete_ByEventID tests that an author can delete own events by id
*/
func TestDelete_ByEventID(t *testing.T) {
	db := useIntegrationDB(t)
	a := &EventAdapter{}
	pk := newPubkey(t, db)
	toDelete := createEvent(t, a, pk, 1, 100, nil)
	toKeep := createEvent(t, a, pk, 1, 200, nil)

	err, allPermitted := a.Delete(deletionRequest(pk, 300, nostr.Tag{"e", toDelete.ID}))
	if err != nil || !allPermitted {
		t.Fatalf("expected a permitted deletion but got %v %v", err, allPermitted)
	}

	remaining := remainingIDs(t, a, pk)
	if remaining[toDelete.ID] || !remaining[toKeep.ID] {
		t.Fatalf("unexpected remaining events: %v", remaining)
	}
}

/*
TestDelete_OtherAuthorsEventByID tests that events of other authors are not deleted by id
*/
func TestDelete_OtherAuthorsEventByID(t *testing.T) {
	db := useIntegrationDB(t)
	a := &EventAdapter{}
	victim := newPubkey(t, db)
	attacker := newPubkey(t, db)
	event := createEvent(t, a, victim, 1, 100, nil)

	if err, _ := a.Delete(deletionRequest(attacker, 300, nostr.Tag{"e", event.ID})); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !remainingIDs(t, a, victim)[event.ID] {
		t.Fatalf("the event of another author must not be deleted")
	}
}

/*
TestDelete_NonExistingEvent tests that deleting an unknown event succeeds
*/
func TestDelete_NonExistingEvent(t *testing.T) {
	db := useIntegrationDB(t)
	a := &EventAdapter{}
	pk := newPubkey(t, db)

	err, allPermitted := a.Delete(deletionRequest(pk, 300, nostr.Tag{"e", nostr.GeneratePrivateKey()}))
	if err != nil || !allPermitted {
		t.Fatalf("expected a successful deletion but got %v %v", err, allPermitted)
	}
}

/*
TestDelete_AddressableUntil tests that an address deletion removes the versions up to the deletion only
*/
func TestDelete_AddressableUntil(t *testing.T) {
	db := useIntegrationDB(t)
	a := &EventAdapter{}
	pk := newPubkey(t, db)
	older := createEvent(t, a, pk, 30023, 100, nostr.Tags{{"d", "article"}})
	newer := createEvent(t, a, pk, 30023, 300, nostr.Tags{{"d", "article"}})
	otherIdentifier := createEvent(t, a, pk, 30023, 100, nostr.Tags{{"d", "another"}})

	err, allPermitted := a.Delete(deletionRequest(pk, 200, nostr.Tag{"a", fmt.Sprintf("30023:%s:article", pk)}))
	if err != nil || !allPermitted {
		t.Fatalf("expected a permitted deletion but got %v %v", err, allPermitted)
	}

	remaining := remainingIDs(t, a, pk)
	if remaining[older.ID] {
		t.Fatalf("the version before the deletion should be deleted")
	}
	if !remaining[newer.ID] {
		t.Fatalf("the version after the deletion must be kept")
	}
	if !remaining[otherIdentifier.ID] {
		t.Fatalf("events with another d identifier must be kept")
	}
}

/*
TestDelete_ReplaceableAddress tests that an address with an empty identifier deletes a replaceable event
*/
func TestDelete_ReplaceableAddress(t *testing.T) {
	db := useIntegrationDB(t)
	a := &EventAdapter{}
	pk := newPubkey(t, db)
	event := createEvent(t, a, pk, 10002, 100, nil)

	if err, _ := a.Delete(deletionRequest(pk, 200, nostr.Tag{"a", fmt.Sprintf("10002:%s:", pk)})); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if remainingIDs(t, a, pk)[event.ID] {
		t.Fatalf("the replaceable event should be deleted")
	}
}

/*
TestDelete_MixedRequest tests that permitted tags are applied even when another tag is not permitted
*/
func TestDelete_MixedRequest(t *testing.T) {
	db := useIntegrationDB(t)
	a := &EventAdapter{}
	pk := newPubkey(t, db)
	other := newPubkey(t, db)
	own := createEvent(t, a, pk, 1, 100, nil)
	foreign := createEvent(t, a, other, 30023, 100, nostr.Tags{{"d", "article"}})

	err, allPermitted := a.Delete(deletionRequest(pk, 200,
		nostr.Tag{"a", fmt.Sprintf("30023:%s:article", other)},
		nostr.Tag{"e", own.ID},
	))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allPermitted {
		t.Fatalf("the request references another author and should not be fully permitted")
	}
	if remainingIDs(t, a, pk)[own.ID] {
		t.Fatalf("the own event should be deleted although another tag was not permitted")
	}
	if !remainingIDs(t, a, other)[foreign.ID] {
		t.Fatalf("the event of another author must not be deleted")
	}
}

/*
TestDelete_MalformedTags tests that malformed tags are ignored and do not delete anything
*/
func TestDelete_MalformedTags(t *testing.T) {
	db := useIntegrationDB(t)
	a := &EventAdapter{}
	pk := newPubkey(t, db)
	profile := createEvent(t, a, pk, 0, 100, nil)

	err, _ := a.Delete(deletionRequest(pk, 200,
		nostr.Tag{"e"},
		nostr.Tag{"a"},
		nostr.Tag{"a", "1"},
		nostr.Tag{"a", fmt.Sprintf("abc:%s:x", pk)},
		nostr.Tag{"a", fmt.Sprintf("abc:%s:", pk)},
	))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !remainingIDs(t, a, pk)[profile.ID] {
		t.Fatalf("malformed tags must not delete the profile event")
	}
}

/*
TestDeleteAndInsertKind3 tests that a new follow list replaces the old one, also when there was none before
*/
func TestDeleteAndInsertKind3(t *testing.T) {
	db := useIntegrationDB(t)
	a := &EventAdapter{}
	pk := newPubkey(t, db)

	first := nostr.Event{ID: nostr.GeneratePrivateKey(), PubKey: pk, CreatedAt: 100, Kind: 3, Tags: nostr.Tags{{"p", "one"}}}
	if err := a.DeleteAndInsertKind3(first); err != nil {
		t.Fatalf("the first follow list should be stored but got %v", err)
	}

	second := nostr.Event{ID: nostr.GeneratePrivateKey(), PubKey: pk, CreatedAt: 200, Kind: 3, Tags: nostr.Tags{{"p", "two"}}}
	if err := a.DeleteAndInsertKind3(second); err != nil {
		t.Fatalf("the second follow list should be stored but got %v", err)
	}

	remaining := remainingIDs(t, a, pk)
	if len(remaining) != 1 || !remaining[second.ID] {
		t.Fatalf("only the newest follow list should remain but got %v", remaining)
	}
}

/*
TestLogAdapterCreate tests that log entries are stored
*/
func TestLogAdapterCreate(t *testing.T) {
	db := useIntegrationDB(t)
	ip := "test-" + nostr.GeneratePrivateKey()[:16]
	t.Cleanup(func() { db.Unscoped().Where("ip = ?", ip).Delete(&models.Log{}) })

	if _, err := (&LogAdapter{}).Create(ip, "TEST", "content"); err != nil {
		t.Fatalf("failed to create log entry: %v", err)
	}

	var count int64
	db.Model(&models.Log{}).Where("ip = ? AND type = ?", ip, "TEST").Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 log entry but got %v", count)
	}
}
