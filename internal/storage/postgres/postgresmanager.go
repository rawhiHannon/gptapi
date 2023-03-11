package postgres

import (
	// "context"
	// "database/sql"
	"errors"
	"log"

	// "os"
	// "strings"
	"sync"
	"time"

	// _ "github.com/lib/pq"
	"bytes"

	"github.com/go-pg/pg/v9"
)

type ReadClient struct {
	client       *pg.DB
	name         string
	timerChannel chan struct{}
	clientLock   *sync.RWMutex
}

func NewReadClient(name string, client *pg.DB, callback func(*ReadClient)) *ReadClient {
	readClient := &ReadClient{}
	readClient.timerChannel = make(chan struct{}, 0)
	readClient.clientLock = new(sync.RWMutex)
	readClient.client = client
	readClient.name = name

	go func() {
		for {
			select {
			case <-readClient.timerChannel:
				continue
			case <-time.After(10 * time.Minute):
				callback(readClient)
				readClient.close()
				return
			}
		}
	}()
	return readClient
}

func (r *ReadClient) close() {
	r.clientLock.Lock()
	defer r.clientLock.Unlock()
	log.Println("ReadClient " + r.name + " closed")
	if r.client != nil {
		r.client.Close()
		r.client = nil
	}
}

func (r *ReadClient) copyTo(query string, buffer *bytes.Buffer) error {
	r.clientLock.RLock()
	defer r.clientLock.RUnlock()
	if r.client == nil {
		//TODO: use error codes
		return errors.New("no client")
	}
	r.resetTimer()
	_, err := r.client.CopyTo(buffer, query)
	if err != nil {
		log.Println(query, err.Error())
		return err
	}
	return nil
}

func (r *ReadClient) query(query string, model interface{}) error {
	r.clientLock.RLock()
	defer r.clientLock.RUnlock()
	if r.client == nil {
		//TODO: use error codes
		return errors.New("no client")
	}
	r.resetTimer()
	_, err := r.client.Query(model, query)
	if err != nil {
		log.Println(query, err.Error())
		return err
	}
	return nil
}

func (r *ReadClient) resetTimer() {
	oldChannel := r.timerChannel
	r.timerChannel = make(chan struct{}, 0)
	close(oldChannel)
}

type PostgresManager struct {
	client *pg.DB
	dbLock *sync.RWMutex
	ready  chan struct{}

	readClientsLock *sync.RWMutex
	readClients     map[string]*ReadClient

	host     string
	port     string
	dbname   string
	user     string
	password string
}

func New(host string, port string, user string, password string) *PostgresManager {
	instance := &PostgresManager{}
	instance.dbLock = new(sync.RWMutex)
	instance.ready = make(chan struct{})

	instance.readClientsLock = new(sync.RWMutex)
	instance.readClients = make(map[string]*ReadClient)

	instance.host = host
	instance.port = port
	instance.dbname = "postgres"
	instance.user = user
	instance.password = password

	instance.connect(instance.dbname)
	return instance
}

func (p *PostgresManager) setReady(status bool) {
	if p.ready == nil {
		p.ready = make(chan struct{})
	}
	select {
	case <-p.ready:
		if status == true {
			return
		} else {
			p.ready = make(chan struct{})
		}
	default:
		if status == false {
			return
		} else {
			close(p.ready)
		}
	}
}

func (p *PostgresManager) isDatabaseExists(dbname string) bool {
	dbList := p.DatabaseList()
	for _, name := range dbList {
		if name == dbname {
			return true
		}
	}
	return false
}

// TODO: handle connection fail with ping pong
func (p *PostgresManager) connect(dbname string) {
	if dbname == "" {
		dbname = "postgres"
	}
	if p.client != nil {
		p.client.Close()
	}
	p.dbLock.Lock()
	defer p.dbLock.Unlock()
	p.client = pg.Connect(&pg.Options{
		User:     p.user,
		Password: p.password,
		Database: dbname,
		Addr:     p.host + ":" + p.port,
	})
	p.dbname = dbname
	p.setReady(true)
}

func (p *PostgresManager) getReadClient(dbName string) *ReadClient {
	p.readClientsLock.RLock()
	if readClient, exists := p.readClients[dbName]; exists {
		p.readClientsLock.RUnlock()
		return readClient
	}
	p.readClientsLock.RUnlock()
	p.readClientsLock.Lock()
	defer p.readClientsLock.Unlock()
	if readClient, exists := p.readClients[dbName]; exists {
		return readClient
	} else {
		client := pg.Connect(&pg.Options{
			User:     p.user,
			Password: p.password,
			Database: dbName,
			Addr:     p.host + ":" + p.port,
		})
		readClient := NewReadClient(dbName, client, func(rclient *ReadClient) {
			p.readClientsLock.Lock()
			defer p.readClientsLock.Unlock()
			delete(p.readClients, dbName)
		})

		p.readClients[dbName] = readClient
		return readClient
	}
}

func (p *PostgresManager) SwitchDatabase(dbname string) error {
	p.dbLock.RLock()
	if p.dbname == dbname {
		p.dbLock.RUnlock()
		return nil
	}
	p.dbLock.RUnlock()
	if p.isDatabaseExists(dbname) == true {
		p.connect(dbname)
		sqlQuery := `GRANT SELECT ON ALL TABLES IN SCHEMA public TO support;
		ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO support;
		GRANT SELECT ON ALL TABLES IN SCHEMA public TO customer;
		ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO customer;
		GRANT SELECT ON ALL TABLES IN SCHEMA public TO platis;
		ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO platis;`
		err := p.RawQuery(sqlQuery)
		if err != nil {
			return err
		}
		return nil
	}
	//TODO: maybe create it ?
	return errors.New(dbname + " database not found")
}

func (p *PostgresManager) DatabaseList() []string {
	sqlQuery := `SELECT datname FROM pg_database WHERE datistemplate = false;`
	dbList := make([]string, 0)
	err := p.Query(&dbList, sqlQuery)
	if err != nil {
		return []string{}
	}
	return dbList
}

// TODO: add logs
func (p *PostgresManager) CreateDatabase(dbname string) error {
	if p.client == nil {
		return errors.New("Not connected")
	}
	if p.isDatabaseExists(dbname) == true {
		return errors.New(dbname + " database already exists")
	}
	sqlQuery := `CREATE DATABASE ` + dbname

	err := p.RawQuery(sqlQuery)
	if err != nil {
		return err
		//TODO:
	}
	return nil
}

func (p *PostgresManager) Close() {
	p.dbLock.Lock()
	p.setReady(false)
	p.client.Close()
	p.dbLock.Unlock()
	log.Println("Postgres successfully terminated!")
}

func (p *PostgresManager) RawQuery(sqlStatement string) error {
	// return nil
	<-p.ready
	p.dbLock.RLock()
	_, err := p.client.Exec(sqlStatement)
	p.dbLock.RUnlock()
	if err != nil {
		log.Println(sqlStatement, err.Error())
		return err
	}
	return nil
}

// TODO: change from buffer to writer to support all types
func (p *PostgresManager) CopyTo(dbName string, query string, buffer *bytes.Buffer) error {
	client := p.getReadClient(dbName)
	retries := 0

	for {
		if retries > 2 {
			//TODO: use error code
			return errors.New("")
		}
		retries++
		err := client.copyTo(query, buffer)
		//TODO: recognize regular errors
		if err != nil {
			continue
		}
		return nil
	}
}

// TODO: add other details like results count
func (p *PostgresManager) QueryDatabase(model interface{}, dbName string, query string) error {
	client := p.getReadClient(dbName)
	retries := 0

	for {
		if retries > 2 {
			//TODO: use error code
			return errors.New("")
		}
		retries++
		err := client.query(query, model)
		//TODO: recognize regular errors
		if err != nil {
			continue
		}
		return nil
	}
}

// TODO: add other details like results count
func (p *PostgresManager) Query(model interface{}, sqlStatement string) error {
	<-p.ready
	p.dbLock.RLock()
	_, err := p.client.Query(model, sqlStatement)
	p.dbLock.RUnlock()
	if err != nil {
		log.Println(sqlStatement, err.Error())
		return err
	}
	return nil
}
