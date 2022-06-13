package eth

import (
	"fmt"
	"log"
	"sync"
)

type Fdb struct {
	entries map[string]Interface
	l       sync.RWMutex
}

func (db *Fdb) Add(ifce Interface, addrs ...MAC) {
	db.l.Lock()
	defer db.l.Unlock()

	if db.entries == nil {
		db.entries = map[string]Interface{}
	}
	ch := false
	for _, addr := range addrs {
		if addr.IsZero() || addr.IsBroadcast() {
			return
		}
		saddr := addr.String()

		if db.entries[saddr] != ifce {
			ch = true
		}
		db.entries[saddr] = ifce
	}
	if ch {
		log.Printf("fdb changed (add):\n:%s", db.print())
	}
}

func (db *Fdb) Get(addr MAC) (ifce Interface, ok bool) {
	db.l.Lock()
	defer db.l.Unlock()
	if db.entries != nil {
		ifce, ok = db.entries[addr.String()]
	}
	return
}

func (db *Fdb) Remove(addr MAC) {
	db.l.Lock()
	defer db.l.Unlock()
	if db.entries == nil {
		return
	}
	delete(db.entries, addr.String())
	log.Printf("fdb changed (remove):\n:%s", db.print())
}

func (db *Fdb) Clear(ifce Interface) {
	db.l.Lock()
	defer db.l.Unlock()
	if db.entries == nil {
		return
	}
	for addr, i := range db.entries {
		if i == ifce {
			delete(db.entries, addr)
		}
	}
	log.Printf("fdb changed (clear):\n:%s", db.print())
}

func (db *Fdb) Flush() {
	db.l.Lock()
	defer db.l.Unlock()
	db.entries = nil
	log.Printf("fdb changed (flush):\n:%s", db.print())
}

func (db *Fdb) print() string {
	str := ""
	for addr, ifce := range db.entries {
		str += fmt.Sprintf("\t%s\t%s\n", addr, ifce.Name())
	}
	return str
}
