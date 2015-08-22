package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/robfig/cron"
)

var keeper *Keeper

type Keeper struct {
	cr      *cron.Cron
	crmu    sync.Mutex
	tkmu    sync.RWMutex
	tasks   map[string]Task
	runRecs map[string]*Record
}

func NewKeeper(tasks []Task) *Keeper {
	k := &Keeper{
		cr:      cron.New(),
		tasks:   make(map[string]Task, 0),
		runRecs: make(map[string]*Record, 0),
	}
	for _, task := range tasks {
		ta := task
		if err := k.AddTask(ta); err != nil {
			log.Println(err)
			continue
		}
		taskFunc := func() {
			if err := ta.Run(TRIGGER_SCHEDULE); err != nil {
				log.Println(ta.Name, err)
			}
		}
		k.cr.AddFunc(task.Schedule, taskFunc)
	}
	return k
}

func (k *Keeper) NewRecord(name string) (key string, rec *Record, err error) {
	k.tkmu.RLock()
	defer k.tkmu.RUnlock()

	//log.Println(k.tasks)
	if _, ok := k.tasks[name]; !ok {
		return "", nil, fmt.Errorf("No such task has name: %s", name)
	}

	var count int
	//db.Model(Record{}).Where("name = ?", name).Count(&count)
	idx := count
	rec = &Record{
		Name:   name,
		Index:  idx,
		Buffer: bytes.NewBuffer(nil),
	}
	key = rec.Key()
	k.crmu.Lock()
	k.runRecs[key] = rec
	k.crmu.Unlock()
	return key, rec, nil
}

func (k *Keeper) DoneRecord(key string) error {
	k.crmu.Lock()
	defer k.crmu.Unlock()
	if rec, ok := k.runRecs[key]; ok {
		if err := rec.Done(); err != nil {
			return err
		}
		delete(k.runRecs, key)
		return nil
	}
	return errors.New("Record not found in keeper")
}

func (k *Keeper) ListRecords(limit int) (rs []*Record, err error) {
	rs = make([]*Record, 0)
	for _, rec := range k.runRecs {
		rs = append(rs, rec)
	}
	// Need to find in db
	return rs, nil
}

func (k *Keeper) Reload() {
	k.cr.Stop()
	k.cr.Start()
}

func (k *Keeper) AddTask(t Task) error {
	k.tkmu.Lock()
	defer k.tkmu.Unlock()
	if _, exists := k.tasks[t.Name]; exists {
		return errors.New("Task name duplicate: " + t.Name)
	}
	k.tasks[t.Name] = t
	return nil
}

func (k *Keeper) DelTask(name string) {
	k.tkmu.Lock()
	defer k.tkmu.Unlock()
	delete(k.tasks, name)
}

func (k *Keeper) PutTask(name string, t Task) error {
	if name != t.Name {
		return errors.New("Task name not correct")
	}
	k.tkmu.Lock()
	defer k.tkmu.Unlock()
	k.tasks[name] = t
	return nil
}

func (k *Keeper) Tasks() []Task {
	var ts = make([]Task, 0)
	for _, task := range k.tasks {
		ts = append(ts, task)
	}
	return ts
}
