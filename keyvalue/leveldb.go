package keyvalue

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func OpenLevelDB(path string) (DB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &levelDB{db: db}, nil
}

type levelDB struct {
	db *leveldb.DB
}

func (l *levelDB) Get(key string) (value []byte, err error) {
	value, err = l.db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		err = ErrNotFound
	}
	return value, err
}

func (l *levelDB) Set(key string, value []byte) error {
	return l.db.Put([]byte(key), value, nil)
}

func (l *levelDB) Delete(key string) error {
	err := l.db.Delete([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		err = nil
	}
	return err
}

func (l *levelDB) Close() error {
	return l.db.Close()
}

func (l *levelDB) Iterator(start, end string) Iterator {
	var e []byte
	if end != "" {
		e = []byte(end)
	}
	return &ldbIt{
		l.db.NewIterator(&util.Range{
			Start: []byte(start),
			Limit: e,
		}, nil)}
}

type ldbIt struct {
	it iterator.Iterator
}

func (it *ldbIt) Next() bool {
	return it.it.Next()
}

func (it *ldbIt) Key() string {
	return string(it.it.Key())
}

func (it *ldbIt) Value() []byte {
	return it.it.Value()
}

func (it *ldbIt) Err() error {
	return it.it.Error()
}

func (it *ldbIt) Close() error {
	it.it.Release()
	return nil
}

func (l *levelDB) Batch() Batch {
	return &lbatch{db: l.db}
}

type lbatch struct {
	leveldb.Batch
	db *leveldb.DB
}

func (l *lbatch) Set(key string, value []byte) {
	l.Batch.Put([]byte(key), value)
}

func (l *lbatch) Delete(key string) {
	l.Batch.Delete([]byte(key))
}

func (l *lbatch) Commit() error {
	return l.db.Write(&l.Batch, nil)
}
