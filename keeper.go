package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/qiniu/log"

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
		//cr:      cron.New(),
		tasks:   make(map[string]Task, 0),
		runRecs: make(map[string]*Record, 0),
	}
	for _, t := range tasks {
		k.tasks[t.Name] = t
	}
	k.reloadCron()
	return k
}

func (k *Keeper) reloadCron() {
	if k.cr != nil {
		k.cr.Stop()
	}
	k.cr = cron.New()
	for _, task := range k.Tasks() {
		ta := task
		//log.Println(ta, ta.Enabled)
		if !ta.Enabled {
			continue
		}
		taskFunc := func() {
			log.Println(ta)
			if err := ta.Run(TRIGGER_SCHEDULE); err != nil {
				log.Println(ta.Name, err)
			}
		}
		k.cr.AddFunc(task.Schedule, taskFunc)
	}
	k.cr.Start()
}

func (k *Keeper) NewRecord(name string) (key string, rec *Record, err error) {
	k.tkmu.RLock()
	defer k.tkmu.RUnlock()

	task, ok := k.tasks[name]
	if !ok {
		return "", nil, fmt.Errorf("No such task has name: %s", name)
	}

	var count int
	//cnt, err := xe.Where("name = ?", name).Count(Record{})
	//log.Println(cnt, err)
	//db.Model(Record{}).Where("name = ?", name).Count(&count)
	idx := count
	rec = &Record{
		Name:   name,
		Index:  idx,
		Buffer: bytes.NewBuffer(nil),
		T:      task,
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
	k.reloadCron()
	//k.cr.Stop()
	//k.cr.Start()
}

func (k *Keeper) AddTask(t Task) error {
	k.tkmu.Lock()
	defer k.tkmu.Unlock()
	if _, exists := k.tasks[t.Name]; exists {
		return errors.New("Task name duplicate: " + t.Name)
	}
	if _, err := cron.Parse(t.Schedule); err != nil {
		return err
	}
	t.Enabled = true
	k.tasks[t.Name] = t
	k.reloadCron()
	return k.Save()
}

func (k *Keeper) DelTask(name string) error {
	k.tkmu.Lock()
	defer k.tkmu.Unlock()
	delete(k.tasks, name)
	k.reloadCron()
	return k.Save()
}

func (k *Keeper) PutTask(name string, t Task) error {
	if name != t.Name {
		return errors.New("Task name not correct")
	}
	if _, err := cron.Parse(t.Schedule); err != nil {
		return err
	}
	k.tkmu.Lock()
	k.tasks[name] = t
	k.tkmu.Unlock()
	k.reloadCron()
	return k.Save()
}

func (k *Keeper) Tasks() []Task {
	//k.tkmu.RLock()
	//defer k.tkmu.RUnlock()

	var ts = make([]Task, 0)
	for _, task := range k.tasks {
		ts = append(ts, task)
	}
	return ts
}

func (k *Keeper) Save() error {
	data, err := json.MarshalIndent(k.Tasks(), "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(gcfg.SchedFile, data, 0644)
}
