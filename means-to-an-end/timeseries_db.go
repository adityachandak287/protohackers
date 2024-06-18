package main

import (
	"fmt"
	"math"
	"slices"
	"sync"
)

type TSDB interface {
	RegisterId(id string) error
	UnregisterId(id string)
	Insert(id string, timestamp int32, price int32) error
	Query(id string, startTs int32, endTs int32) ([]int32, error)
	QueryAvg(id string, startTs int32, endTs int32) (int32, error)
}

type ReadingsMap = map[int32]int32

type MapTSDB struct {
	data map[string]ReadingsMap
	lock sync.Mutex
}

func NewMapTSDB() *MapTSDB {
	return &MapTSDB{
		data: make(map[string]ReadingsMap),
	}
}

func (db *MapTSDB) RegisterId(id string) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	_, exists := db.data[id]

	if exists {
		return fmt.Errorf("id %s already registered", id)
	}
	db.data[id] = make(ReadingsMap)
	return nil
}

func (db *MapTSDB) UnregisterId(id string) {
	db.lock.Lock()
	defer db.lock.Unlock()

	_, exists := db.data[id]

	if exists {
		delete(db.data, id)
	}
}

func (db *MapTSDB) Insert(id string, timestamp int32, price int32) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.data[id][timestamp] = price
	return nil
}

func (db *MapTSDB) Query(id string, startTs int32, endTs int32) ([]int32, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	data, exists := db.data[id]
	if !exists {
		return nil, fmt.Errorf("id %s not registered", id)
	}

	timestamps := make([]int32, 0, len(data))
	for ts := range data {
		timestamps = append(timestamps, ts)
	}
	slices.Sort(timestamps)

	values := make([]int32, 0, len(data))
	for _, ts := range timestamps {
		if ts >= startTs && ts <= endTs {
			values = append(values, data[ts])
		}
	}
	return values, nil
}

func (db *MapTSDB) QueryAvg(id string, startTs int32, endTs int32) (int32, error) {
	values, err := db.Query(id, startTs, endTs)

	if err != nil {
		return 0, err
	}

	total := int64(0)
	for _, value := range values {
		total += int64(value)
	}

	if len(values) == 0 {
		return 0, nil
	}

	return int32(math.Round(float64(total / int64(len(values))))), nil
}
