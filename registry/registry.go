package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type Instance struct {
	Env             string   `json:"env"`
	AppID           string   `json:"appid"`
	Hostname        string   `json:"hostname"`
	Addrs           []string `json:"addrs"`
	Version         string   `json:"version"`
	Status          uint32   `json:"status"`
	RegTimestamp    int64    `json:"reg_timestamp"`
	UpTimestamp     int64    `json:"up_timestamp"`
	RenewTimestamp  int64    `json:"renew_timestamp"`
	DirtyTimestamp  int64    `json:"dirty_timestamp"`
	LatestTimestamp int64    `json:"latest_timestamp"`
}

type Application struct {
	appid           string
	instances       map[string]*Instance
	latestTimestamp int64
	lock            sync.RWMutex
}

type Registry struct {
	apps map[string]*Application
	lock sync.RWMutex
}

type RequestRegister struct {
	Env             string
	AppId           string
	Hostname        string
	Addrs           []string
	Status          uint32
	Version         string
	LatestTimestamp int64
}

func NewRegistry() *Registry {
	registry := &Registry{
		apps: make(map[string]*Application),
	}
	return registry
}

func NewInstance(req *RequestRegister) *Instance {
	now := time.Now().UnixNano()
	instance := &Instance{
		Env:             req.Env,
		AppID:           req.AppId,
		Hostname:        req.Hostname,
		Addrs:           req.Addrs,
		Version:         req.Version,
		Status:          req.Status,
		RegTimestamp:    now,
		UpTimestamp:     now,
		RenewTimestamp:  now,
		DirtyTimestamp:  now,
		LatestTimestamp: now,
	}
	return instance
}

func NewApplication(appid string) *Application {
	return &Application{
		appid:     appid,
		instances: make(map[string]*Instance),
	}
}

// 服务注册
func (r *Registry) Register(instance *Instance, latestTimestamp int64) (*Application, error) {
	key := getKey(instance.AppID, instance.Env)
	r.lock.RLock()
	app, ok := r.apps[key]
	r.lock.RUnlock()
	if !ok {
		app = NewApplication(instance.AppID)
	}
	_, isNew := app.AddInstance(instance, latestTimestamp)
	if isNew {
		// TODO
	}

	log.Println("action register...")
	r.lock.Lock()
	r.apps[key] = app
	r.lock.Unlock()
	return app, nil
}

func (app *Application) AddInstance(in *Instance, latestTimestamp int64) (*Instance, bool) {
	app.lock.Lock()
	defer app.lock.Unlock()
	appIns, ok := app.instances[in.Hostname]
	if ok {
		in.UpTimestamp = appIns.UpTimestamp
		if in.DirtyTimestamp < appIns.DirtyTimestamp {
			log.Println("register exist dirty timestamp")
			in = appIns
		}
	}
	app.instances[in.Hostname] = in
	app.upLatestTimestamp(latestTimestamp)
	returnIns := new(Instance)
	*returnIns = *in
	return returnIns, !ok
}

func (app *Application) upLatestTimestamp(latestTimestamp int64) {
	app.latestTimestamp = latestTimestamp
}

func getKey(appid string, env string) string {
	return fmt.Sprintf("%s:%s", appid, env)
}

type FetchData struct {
	Instances       []*Instance
	LatestTimestamp int64
}

func (r *Registry) getApplication(appid, env string) (*Application, bool) {
	key := getKey(appid, env)
	r.lock.RLock()
	app, ok := r.apps[key]
	r.lock.RUnlock()
	return app, ok
}

// 服务发现
func (r *Registry) Fetch(env, appid string, status uint32, latestTime int64) (*FetchData, error) {
	app, ok := r.getApplication(appid, env)
	if !ok {
		return nil, errors.New("NotFound")
	}
	return app.GetInstance(status, latestTime)
}

func (app *Application) GetInstance(status uint32, latestTime int64) (*FetchData, error) {
	app.lock.RLock()
	defer app.lock.RUnlock()
	fmt.Println(latestTime, app.latestTimestamp)
	if latestTime >= app.latestTimestamp {
		return nil, errors.New("NotModified")
	}
	fetchData := FetchData{
		Instances:       make([]*Instance, 0),
		LatestTimestamp: app.latestTimestamp,
	}
	var exists bool
	for _, instance := range app.instances {
		if status&instance.Status > 0 {
			exists = true
			newInstance := copyInstance(instance)
			fetchData.Instances = append(fetchData.Instances, newInstance)
		}
	}
	if !exists {
		return nil, errors.New("NotFound")
	}
	return &fetchData, nil
}

func copyInstance(src *Instance) *Instance {
	dst := new(Instance)
	*dst = *src
	dst.Addrs = make([]string, len(src.Addrs))
	for i, addr := range src.Addrs {
		dst.Addrs[i] = addr
	}
	return dst
}

// 服务下线
func (r *Registry) Cancel(env, appid, hostname string, latestTimestamp int64) (*Instance, error) {
	log.Println("action cancel...")
	app, ok := r.getApplication(appid, env)
	if !ok {
		return nil, errors.New("NotFound")
	}
	instance, ok, insLen := app.Cancel(hostname, latestTimestamp)
	if !ok {
		return nil, errors.New("NotFound")
	}
	if insLen == 0 {
		r.lock.Lock()
		delete(r.apps, getKey(appid, env))
		r.lock.Unlock()
	}
	return instance, nil
}

func (app *Application) Cancel(hostname string, latestTimestamp int64) (*Instance, bool, int) {
	newInstance := new(Instance)
	app.lock.Lock()
	defer app.lock.Unlock()
	appIn, ok := app.instances[hostname]
	if !ok {
		return nil, ok, 0
	}
	delete(app.instances, hostname)
	appIn.LatestTimestamp = latestTimestamp
	app.upLatestTimestamp(latestTimestamp)
	*newInstance = *appIn
	return newInstance, true, len(app.instances)
}

// 服务续约
func (r *Registry) Renew(env, appid, hostname string) (*Instance, error) {
	app, ok := r.getApplication(appid, env)
	if !ok {
		return nil, errors.New("NotFound")
	}
	in, ok := app.Renew(hostname)
	if !ok {
		return nil, errors.New("NotFound")
	}
	return in, nil
}

func (app *Application) Renew(hostname string) (*Instance, bool) {
	app.lock.Lock()
	defer app.lock.Unlock()
	appIn, ok := app.instances[hostname]
	if !ok {
		return nil, ok
	}
	appIn.RenewTimestamp = time.Now().UnixNano()
	return copyInstance(appIn), true
}

const (
	CheckEvictInterval = 60
)

func (r *Registry) evictTask()  {
	ticker := time.Tick(CheckEvictInterval)
	for {
		select {
		case <-ticker:
			r.evict()
		}
	}
}

func (r *Registry) evict()  {

}

func (r *Registry) getAllApplications()  {

}

func main() {
	var req = &RequestRegister{
		Env:             "dev",
		AppId:           "com.xx.testapp",
		Hostname:        "myhost",
		Addrs:           []string{"http://testapp.xx.com/myhost"},
		Status:          1,
		LatestTimestamp: time.Now().UnixNano(),
	}
	r := NewRegistry()
	instance := NewInstance(req)
	r.Register(instance, req.LatestTimestamp)
	fetchData, err := r.Fetch(req.Env, req.AppId, req.Status, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	r.Cancel(req.Env, req.AppId, req.Hostname, 0)
	log.Println(fetchData)
}
