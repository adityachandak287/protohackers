package main

import (
	"fmt"
	"testing"
)

func TestMapTSDB(t *testing.T) {
	db := NewMapTSDB()

	id := "id1"
	db.RegisterId(id)

	db.Insert(id, 12345, 101)
	db.Insert(id, 12346, 102)
	db.Insert(id, 12347, 100)
	db.Insert(id, 40960, 5)

	avg, err := db.QueryAvg(id, 12288, 16384)
	if err != nil {
		t.Fatalf("Error while querying average: %s", err)
	}

	expectedAvg := int32(101)
	if avg != expectedAvg {
		t.Fatalf("Incorrect average! Expected %d != Output %d", expectedAvg, avg)
	}

	db.UnregisterId(id)
}

func TestMapTSDBParallel(t *testing.T) {
	db := NewMapTSDB()

	for idx := 0; idx < 10; idx++ {
		mapTSDBParallelIter(t, db, int32(idx))
	}
}

func mapTSDBParallelIter(t *testing.T, db TSDB, idx int32) {
	t.Run("Verify MapTSDB Thread Safety", func(t *testing.T) {
		t.Parallel()
		id := fmt.Sprintf("id%d", idx)
		db.RegisterId(id)

		ts := int32(10000)
		for insertIdx := 0; insertIdx < int(idx); insertIdx++ {
			db.Insert(id, ts, idx)
			ts++
		}

		avg, err := db.QueryAvg(id, ts-idx, ts)
		if err != nil {
			t.Fatalf("Error while querying average: %s", err)
		}

		expectedAvg := int32(idx)
		if avg != expectedAvg {
			t.Fatalf("Incorrect average! Expected %d != Output %d", expectedAvg, avg)
		}

		db.UnregisterId(id)
	})
}
