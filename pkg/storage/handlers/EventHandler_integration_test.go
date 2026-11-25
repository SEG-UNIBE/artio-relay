package handlers

import (
	"testing"

	"github.com/SEG-UNIBE/artio-relay/pkg/config"
	"github.com/SEG-UNIBE/artio-relay/pkg/storage/models"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
newIntegrationHandler connects to the database from .env.tests, the test is skipped when it is not reachable
*/
func newIntegrationHandler(t *testing.T) EventHandler {
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
	return EventHandler{&BaseHandler{Connection: db}}
}

/*
newPubkey returns a random pubkey, every test uses its own pubkey to stay isolated from other data
*/
func newPubkey(t *testing.T, h EventHandler) string {
	t.Helper()
	pubkey := nostr.GeneratePrivateKey()
	t.Cleanup(func() {
		h.Connection.Unscoped().Where("pubkey = ?", pubkey).Delete(&models.Event{})
	})
	return pubkey
}

/*
insertEvent stores an event with the given values and a random id
*/
func insertEvent(t *testing.T, h EventHandler, pubkey string, kind uint32, created int64, tags nostr.Tags) models.Event {
	t.Helper()
	return insertEventWithID(t, h, nostr.GeneratePrivateKey(), pubkey, kind, created, tags)
}

/*
insertEventWithID stores an event with the given values
*/
func insertEventWithID(t *testing.T, h EventHandler, id string, pubkey string, kind uint32, created int64, tags nostr.Tags) models.Event {
	t.Helper()
	if tags == nil {
		tags = nostr.Tags{}
	}
	event := models.Event{EventId: id, Pubkey: pubkey, Kind: kind, Created: created, Tags: tags, Content: "content", Sig: "sig"}
	result, _ := h.CreateEvent(event)
	if err := result.(*gorm.DB).Error; err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}
	return event
}

/*
eventIDs returns the ids of the given events in order
*/
func eventIDs(events []models.Event) []string {
	ids := make([]string, len(events))
	for i, event := range events {
		ids[i] = event.EventId
	}
	return ids
}

/*
expectIDs checks that the events have exactly the expected ids in the expected order
*/
func expectIDs(t *testing.T, events []models.Event, expected ...string) {
	t.Helper()
	got := eventIDs(events)
	if len(got) != len(expected) {
		t.Fatalf("expected %v but got %v", expected, got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("expected %v but got %v", expected, got)
		}
	}
}

/*
getEvents runs GetEvents and fails the test on error
*/
func getEvents(t *testing.T, h EventHandler, filter nostr.Filter) []models.Event {
	t.Helper()
	events, err := h.GetEvents(filter)
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	return events
}

/*
TestGetEvents_NewestFirst tests that events are returned newest first
*/
func TestGetEvents_NewestFirst(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	a := insertEvent(t, h, pk, 1, 100, nil)
	b := insertEvent(t, h, pk, 1, 300, nil)
	c := insertEvent(t, h, pk, 1, 200, nil)

	events := getEvents(t, h, nostr.Filter{Authors: []string{pk}, Limit: 10})
	expectIDs(t, events, b.EventId, c.EventId, a.EventId)
}

/*
TestGetEvents_TieBreakByID tests that events with the same timestamp are ordered by lowest id first
*/
func TestGetEvents_TieBreakByID(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	insertEventWithID(t, h, "b"+pk[1:], pk, 1, 100, nil)
	insertEventWithID(t, h, "a"+pk[1:], pk, 1, 100, nil)

	events := getEvents(t, h, nostr.Filter{Authors: []string{pk}, Limit: 10})
	expectIDs(t, events, "a"+pk[1:], "b"+pk[1:])
}

/*
TestGetEvents_Limit tests that the limit returns the newest events only
*/
func TestGetEvents_Limit(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	insertEvent(t, h, pk, 1, 100, nil)
	b := insertEvent(t, h, pk, 1, 200, nil)
	c := insertEvent(t, h, pk, 1, 300, nil)

	events := getEvents(t, h, nostr.Filter{Authors: []string{pk}, Limit: 2})
	expectIDs(t, events, c.EventId, b.EventId)
}

/*
TestGetEvents_NoLimit tests that a filter without limit returns all events (used by deletions)
*/
func TestGetEvents_NoLimit(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	for i := int64(0); i < 5; i++ {
		insertEvent(t, h, pk, 1, 100+i, nil)
	}

	if events := getEvents(t, h, nostr.Filter{Authors: []string{pk}}); len(events) != 5 {
		t.Fatalf("expected 5 events but got %v", len(events))
	}
}

/*
TestGetEvents_LimitZero tests that an explicit limit of 0 returns no events
*/
func TestGetEvents_LimitZero(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	insertEvent(t, h, pk, 1, 100, nil)

	if events := getEvents(t, h, nostr.Filter{Authors: []string{pk}, LimitZero: true}); len(events) != 0 {
		t.Fatalf("expected no events but got %v", len(events))
	}
}

/*
TestGetEvents_SinceUntil tests that since and until are inclusive bounds
*/
func TestGetEvents_SinceUntil(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	insertEvent(t, h, pk, 1, 100, nil)
	b := insertEvent(t, h, pk, 1, 200, nil)
	c := insertEvent(t, h, pk, 1, 300, nil)
	insertEvent(t, h, pk, 1, 400, nil)

	since, until := nostr.Timestamp(200), nostr.Timestamp(300)
	events := getEvents(t, h, nostr.Filter{Authors: []string{pk}, Since: &since, Until: &until, Limit: 10})
	expectIDs(t, events, c.EventId, b.EventId)
}

/*
TestGetEvents_IDsAndKinds tests the id and kind filters
*/
func TestGetEvents_IDsAndKinds(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	a := insertEvent(t, h, pk, 1, 100, nil)
	b := insertEvent(t, h, pk, 7, 200, nil)
	insertEvent(t, h, pk, 7, 300, nil)

	expectIDs(t, getEvents(t, h, nostr.Filter{IDs: []string{a.EventId, b.EventId}, Limit: 10}), b.EventId, a.EventId)
	expectIDs(t, getEvents(t, h, nostr.Filter{Authors: []string{pk}, Kinds: []int{1}, Limit: 10}), a.EventId)
}

/*
TestGetEvents_TagPosition tests that tag filters only match the tag name at position 0 and the value at position 1
*/
func TestGetEvents_TagPosition(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	match := insertEvent(t, h, pk, 1, 100, nostr.Tags{{"e", "target", "wss://relay"}})
	insertEvent(t, h, pk, 1, 200, nostr.Tags{{"x", "e", "target"}})
	insertEvent(t, h, pk, 1, 300, nostr.Tags{{"target", "e"}})
	insertEvent(t, h, pk, 1, 400, nostr.Tags{{"e", "other"}})

	events := getEvents(t, h, nostr.Filter{Authors: []string{pk}, Tags: nostr.TagMap{"e": {"target"}}, Limit: 10})
	expectIDs(t, events, match.EventId)
}

/*
TestGetEvents_TagValuesOr tests that multiple values of one tag are combined with OR
*/
func TestGetEvents_TagValuesOr(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	a := insertEvent(t, h, pk, 1, 100, nostr.Tags{{"e", "one"}})
	b := insertEvent(t, h, pk, 1, 200, nostr.Tags{{"e", "two"}})
	insertEvent(t, h, pk, 1, 300, nostr.Tags{{"e", "three"}})

	events := getEvents(t, h, nostr.Filter{Authors: []string{pk}, Tags: nostr.TagMap{"e": {"one", "two"}}, Limit: 10})
	expectIDs(t, events, b.EventId, a.EventId)
}

/*
TestGetEvents_TagKeysAnd tests that different tag names are combined with AND
*/
func TestGetEvents_TagKeysAnd(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	both := insertEvent(t, h, pk, 1, 100, nostr.Tags{{"e", "one"}, {"p", "someone"}})
	insertEvent(t, h, pk, 1, 200, nostr.Tags{{"e", "one"}})
	insertEvent(t, h, pk, 1, 300, nostr.Tags{{"p", "someone"}})

	events := getEvents(t, h, nostr.Filter{Authors: []string{pk}, Tags: nostr.TagMap{"e": {"one"}, "p": {"someone"}}, Limit: 10})
	expectIDs(t, events, both.EventId)
}

/*
TestGetEvents_TagFilterBeforeLimit tests that the limit counts matching events only
*/
func TestGetEvents_TagFilterBeforeLimit(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	match := insertEvent(t, h, pk, 1, 100, nostr.Tags{{"e", "target"}})
	for i := int64(0); i < 3; i++ {
		insertEvent(t, h, pk, 1, 200+i, nostr.Tags{{"e", "other"}})
	}

	events := getEvents(t, h, nostr.Filter{Authors: []string{pk}, Tags: nostr.TagMap{"e": {"target"}}, Limit: 1})
	expectIDs(t, events, match.EventId)
}

/*
TestGetEvents_EmptyTagValues tests that a tag filter without values matches nothing
*/
func TestGetEvents_EmptyTagValues(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	insertEvent(t, h, pk, 1, 100, nostr.Tags{{"e", "target"}})

	if events := getEvents(t, h, nostr.Filter{Authors: []string{pk}, Tags: nostr.TagMap{"e": {}}, Limit: 10}); len(events) != 0 {
		t.Fatalf("expected no events but got %v", eventIDs(events))
	}
}

/*
TestDeleteEvents tests that deleted events are no longer returned and that an empty delete succeeds
*/
func TestDeleteEvents(t *testing.T) {
	h := newIntegrationHandler(t)
	pk := newPubkey(t, h)
	a := insertEvent(t, h, pk, 1, 100, nil)
	b := insertEvent(t, h, pk, 1, 200, nil)

	if err := h.DeleteEvents(nil); err != nil {
		t.Fatalf("deleting nothing should succeed but got %v", err)
	}

	toDelete := getEvents(t, h, nostr.Filter{IDs: []string{a.EventId}})
	if err := h.DeleteEvents(toDelete); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	expectIDs(t, getEvents(t, h, nostr.Filter{Authors: []string{pk}, Limit: 10}), b.EventId)
}
