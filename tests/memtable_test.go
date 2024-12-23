package tests

import (
	"fmt"
	"testing"

	swiftdb "github.com/AshikBN/SwiftDB"
	"github.com/stretchr/testify/assert"
)

func TestMemtablePutGetDelete(t *testing.T) {
	t.Parallel()
	memtable := swiftdb.NewMemtable()

	//test put() and get()
	memtable.Put("foo", []byte("bar"))
	result := memtable.Get("foo")
	assert.NotNil(t, result, "memtable.Get(\"foo\") should not return nil")
	assert.Equal(t, []byte("bar"), result.Value, "memtable.Get(\"foo\") should return \"bar\"")
	assert.NotNil(t, result.Timestamp, "memtable.Get(\"foo\") should have a non-nil timestamp")

	// //test delete()
	memtable.Delete("foo")
	assert.Nil(t, memtable.Get("foo").Value, "memtable.Get(\"foo\") should return nil")

	// // get non-existent key
	assert.Nil(t, memtable.Get("abc"), "memtable.Get(\"abc\") should return nil")
}

func TestMemtableScan(t *testing.T) {
	t.Parallel()

	memtable := swiftdb.NewMemtable()

	//test scan with no result
	results := memtable.RangeScan("foo", "foo")
	assert.Empty(t, results, "memtable.Scan(\"foo\", \"foo\") should return an empty slice")

	memtable.Put("foo", []byte("bar"))
	results = memtable.RangeScan("foo", "foo")
	assert.Len(t, results, 1, "memtable.Scan(\"foo\", \"foo\") should return a slice with length 1")
	assert.Equal(t, results[0].Value, []byte("bar"), "memtable.Scan(\"foo\", \"foo\") should return [\"bar\"]")

	memtable.Put("foo", []byte("bar0"))
	memtable.Put("foo8", []byte("bar8"))
	memtable.Put("foo1", []byte("bar1"))
	memtable.Put("foo7", []byte("bar7"))
	memtable.Put("foo3", []byte("bar3"))
	memtable.Put("foo9", []byte("bar9"))
	memtable.Put("foo6", []byte("bar6"))
	memtable.Put("foo2", []byte("bar2"))
	memtable.Put("foo4", []byte("bar4"))
	memtable.Put("foo5", []byte("bar5"))

	results = memtable.RangeScan("foo", "foo9")
	assert.Len(t, results, 10, "memtable.Scan(\"foo\", \"foo9\") should return a slice with length 10")

	for i := 0; i < 10; i++ {
		assert.Equal(t, []byte(fmt.Sprintf("bar%v", i)), results[i].Value, "memtable.Scan(\"foo\", \"foo9\") should return [\"bar0\", \"bar1\", ..., \"bar9\"]")
	}

	// Scan another range
	results = memtable.RangeScan("foo2", "foo7")
	assert.Len(t, results, 6, "memtable.Scan(\"foo2\", \"foo7\") should return a slice with length 6")
	for i := 2; i < 8; i++ {
		assert.Equal(t, []byte(fmt.Sprintf("bar%v", i)), results[i-2].Value, "memtable.Scan(\"foo2\", \"foo7\") should return [\"bar2\", \"bar3\", ..., \"bar7\"]")
	}

	// Scan another range with no results
	results = memtable.RangeScan("foo2", "foo1")
	assert.Empty(t, results, "memtable.Scan(\"foo2\", \"foo1\") should return an empty slice")

	// Scan another range with one result
	results = memtable.RangeScan("foo2", "foo2")
	assert.Len(t, results, 1, "memtable.Scan(\"foo2\", \"foo2\") should return a slice with length 1")
	assert.Equal(t, []byte("bar2"), results[0].Value, "memtable.Scan(\"foo2\", \"foo2\") should return [\"bar2\"]")

	// Scan another range with non-exact start and end keys
	results = memtable.RangeScan("foo2", "fooz")
	assert.Len(t, results, 8, "memtable.Scan(\"foo2\", \"fooz\") should return a slice with length 8")
	for i := 2; i < 10; i++ {
		assert.Equal(t, []byte(fmt.Sprintf("bar%v", i)), results[i-2].Value, "memtable.Scan(\"foo2\", \"fooz\") should return [\"bar2\", \"bar3\", ..., \"bar9\"]")
	}

	// Scan another range with non-exact start and end keys
	results = memtable.RangeScan("a", "foo3")
	assert.Len(t, results, 4, "memtable.Scan(\"fo\", \"foo3\") should return a slice with length 4")
	for i := 0; i < 4; i++ {
		assert.Equal(t, []byte(fmt.Sprintf("bar%v", i)), results[i].Value, "memtable.Scan(\"fo\", \"foo3\") should return [\"bar0\", \"bar1\", \"bar2\", \"bar3\"]")
	}

}

// Test RangeScan with updated entries, run a range scan on updated entries and
// check if the returned entries are correct.
func TestMemtableScanUpdatedEntries(t *testing.T) {
	t.Parallel()

	memtable := swiftdb.NewMemtable()

	// Populate the memtable with a large number of entries
	for i := 0; i < 26; i++ {
		key := fmt.Sprintf("%c", 'a'+i)
		value := []byte(fmt.Sprintf("%c", 'a'+i))
		memtable.Put(key, value)
	}

	// Update some entries
	for i := 0; i < 26; i++ {
		key := fmt.Sprintf("%c", 'a'+i)
		value := []byte(fmt.Sprintf("%c%c", 'a'+i, 'a'+i))
		memtable.Put(key, value)
	}

	// Scan the memtable
	results := memtable.RangeScan("a", "z")

	// Validate the results
	assert.Equal(t, 26, len(results), "Scan results were affected by concurrent operations.")
	for i := 0; i < 26; i++ {
		assert.Equal(t, fmt.Sprintf("%c%c", 'a'+i, 'a'+i), string(results[i].Value), "Scan results were affected by concurrent operations.")
	}
}

// Test the Size() method of the Memtable.
func TestMemtableSize(t *testing.T) {
	t.Parallel()

	memtable := swiftdb.NewMemtable()

	// Test Size() with no entries.
	assert.Equal(t, int64(0), memtable.SizeInBytes(), "memtable.Size() should return 0 with no entries")

	// Test Size() with one entry.
	memtable.Put("foo", []byte("bar"))
	assert.Equal(t, int64(6), memtable.SizeInBytes(), "memtable.Size() should return 6 with one entry")

	// Test Size() with multiple entries.
	memtable.Put("foo", []byte("ba"))
	memtable.Put("foo8", []byte("bar8"))
	memtable.Put("foo1", []byte("bar1"))
	memtable.Put("foo7", []byte("bar7"))
	memtable.Put("foo3", []byte("bar3"))
	memtable.Put("foo9", []byte("bar9"))
	memtable.Put("foo6", []byte("bar6"))
	memtable.Put("foo2", []byte("bar2"))
	memtable.Put("foo4", []byte("bar4"))
	memtable.Put("foo5", []byte("bar5"))
	assert.Equal(t, int64(77), memtable.SizeInBytes(), "memtable.Size() should return 79 with multiple entries")

	// Test Size() with a deleted entry.
	memtable.Delete("foo")
	// We don't subtract the size of the key, since the key remains in the
	// memtable with a tombstone marker.
	assert.Equal(t, int64(75), memtable.SizeInBytes(), "memtable.Size() should return 75 with a deleted entry")
}

// Test GetSerializableEntries() method of the Memtable. Performs a list of put and get operations. Then, it calls
// GetSerializableEntries() and checks if the returned entries are correct.
// The entries should be in sorted order. Deleted entries should be present with a tombstone.
func TestMemtableGetSerializableEntries(t *testing.T) {
	t.Parallel()

	memtable := swiftdb.NewMemtable()

	// Test GetSerializableEntries() with no entries.
	entries := memtable.GetEntries()
	assert.Equal(t, 0, len(entries), "memtable.GetSerializableEntries() should return 0 with no entries")

	// Test GetSerializableEntries() with one entry.
	memtable.Put("foo0", []byte("bar"))
	entries = memtable.GetEntries()
	assert.Equal(t, 1, len(entries), "memtable.GetSerializableEntries() should return 1 with one entry")
	assert.Equal(t, "foo0", entries[0].Key, "memtable.GetSerializableEntries() should return [\"foo\"] with one entry")
	assert.Equal(t, "bar", string(entries[0].Value), "memtable.GetSerializableEntries() should return [\"bar\"] with one entry")
	assert.Equal(t, swiftdb.Command_PUT, entries[0].Command, "memtable.GetSerializableEntries() should return [PUT] with one entry")

	// Test GetSerializableEntries() with multiple entries.
	memtable.Put("foo0", []byte("bar0"))
	memtable.Put("foo8", []byte("bar8"))
	memtable.Put("foo1", []byte("bar1"))
	memtable.Put("foo7", []byte("bar7"))
	memtable.Put("foo3", []byte("bar3"))
	memtable.Put("foo9", []byte("bar9"))
	memtable.Put("foo6", []byte("bar6"))
	memtable.Put("foo2", []byte("bar2"))
	memtable.Put("foo4", []byte("bar4"))
	memtable.Put("foo5", []byte("bar5"))
	entries = memtable.GetEntries()
	assert.Equal(t, 10, len(entries), "memtable.GetSerializableEntries() should return 10 with multiple entries")
	for i := 0; i < 10; i++ {
		assert.Equal(t, fmt.Sprintf("foo%d", i), entries[i].Key, "memtable.GetSerializableEntries() should return [\"foo0\", \"foo1\", ..., \"foo9\"] with multiple entries")
		assert.Equal(t, []byte(fmt.Sprintf("bar%d", i)), entries[i].Value, "memtable.GetSerializableEntries() should return [\"bar0\", \"bar1\", ..., \"bar9\"] with multiple entries")
		assert.Equal(t, swiftdb.Command_PUT, entries[i].Command, "memtable.GetSerializableEntries() should return [PUT] with multiple entries")
	}

	// Test GetSerializableEntries() with a deleted entry.
	memtable.Delete("foo0")
	memtable.Delete("foo8")
	memtable.Delete("foo1")
	// Delete entry not present in memtable.
	memtable.Delete("z")

	entries = memtable.GetEntries()
	assert.Equal(t, 11, len(entries), "memtable.GetSerializableEntries() should return 10 with a deleted entry")
	for i := 0; i < 10; i++ {
		assert.Equal(t, fmt.Sprintf("foo%d", i), entries[i].Key, "memtable.GetSerializableEntries() should return [\"foo0\", \"foo1\", ..., \"foo9\"] with a deleted entry")
		if i == 0 || i == 1 || i == 8 {
			assert.Equal(t, swiftdb.Command_DELETE, entries[i].Command, "memtable.GetSerializableEntries() should return [DELETE] with a deleted entry")
			// Value should be nil.
			assert.Nil(t, entries[i].Value, "memtable.GetSerializableEntries() should return nil with a deleted entry")
		} else {
			assert.Equal(t, swiftdb.Command_PUT, entries[i].Command, "memtable.GetSerializableEntries() should return [PUT] with a deleted entry")
			assert.Equal(t, []byte(fmt.Sprintf("bar%d", i)), entries[i].Value, "memtable.GetSerializableEntries() should return [\"bar0\", \"bar1\", ..., \"bar9\"] with a deleted entry")
		}
	}

	// Check last entry.
	assert.Equal(t, "z", entries[10].Key, "memtable.GetSerializableEntries() should return [\"z\"] with a deleted entry")
}

// Test Clear() method of the Memtable. Performs a list of put and get operations. Then, it calls
// Clear() and checks if the memtable is empty.
func TestMemtableClear(t *testing.T) {
	t.Parallel()

	memtable := swiftdb.NewMemtable()

	// Test Clear() with no entries.
	memtable.Clear()
	assert.Equal(t, int(0), memtable.Len(), "memtable.Len() should return 0 with no entries")
	assert.Equal(t, int64(0), memtable.SizeInBytes(), "memtable.SizeInBytes() should return 0 with no entries")

	// Test Clear() with one entry.
	memtable.Put("foo", []byte("bar"))
	memtable.Clear()
	assert.Equal(t, int(0), memtable.Len(), "memtable.Len() should return 0 with no entries")
	assert.Equal(t, int64(0), memtable.SizeInBytes(), "memtable.SizeInBytes() should return 0 with one entry")

	// Test Clear() with multiple entries.
	memtable.Put("foo0", []byte("bar0"))
	memtable.Put("foo8", []byte("bar8"))
	memtable.Put("foo1", []byte("bar1"))
	memtable.Put("foo7", []byte("bar7"))
	memtable.Put("foo3", []byte("bar3"))
	memtable.Put("foo9", []byte("bar9"))
	memtable.Put("foo6", []byte("bar6"))
	memtable.Put("foo2", []byte("bar2"))
	memtable.Put("foo4", []byte("bar4"))
	memtable.Put("foo5", []byte("bar5"))
	memtable.Clear()
	assert.Equal(t, int(0), memtable.Len(), "memtable.Len() should return 0 with no entries")
	assert.Equal(t, int64(0), memtable.SizeInBytes(), "memtable.SizeInBytes() should return 0 with multiple entries")

	// Test Clear() with a deleted entry.
	memtable.Delete("foo")
	memtable.Clear()
	assert.Equal(t, int(0), memtable.Len(), "memtable.Len() should return 0 with no entries")
	assert.Equal(t, int64(0), memtable.SizeInBytes(), "memtable.SizeInBytes() should return 0 with a deleted entry")
}
