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
		if !ta.Enabled {
			continue
		}
		taskFunc := func() {
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

	total, err := xe.Where("name = ?", name).Count(&Record{})
	// log.Println("NEW REC:", name, total, err)
	idx := int(total)
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

func (k *Keeper) GetRecord(name string, index int) (rec *Record, err error) {
	k.tkmu.RLock()
	defer k.tkmu.RUnlock()
	for _, r := range k.runRecs {
		if r.Name == name && r.Index == index {
			return r, nil
		}
	}
	rec = new(Record)
	xe.ShowSQL = true
	exists, err := xe.Where("`name` = ? AND `index` = ?", name, index).Get(rec)
	xe.ShowSQL = false
	if !exists {
		return nil, fmt.Errorf("Record not exists")
	}
	return rec, nil
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
	var doneRecords []*Record
	if err = xe.Limit(limit).Desc("created_at").Find(&doneRecords); err != nil {
		return nil, err
	}
	for _, rec := range doneRecords {
		rs = append(rs, rec)
	}
	// log.Println(err, doneRecords)
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

// Actually no created
func (k *Keeper) GetOrCreateTask(name string) (task Task, created bool) {
	k.tkmu.Lock()
	defer k.tkmu.Unlock()
	if tk, ok := k.tasks[name]; ok {
		return tk, false
	}
	return Task{
		Name: name,
	}, true
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
