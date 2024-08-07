package cache

import (
	"copilot-gpt4-service/config"
	"copilot-gpt4-service/log"
	"copilot-gpt4-service/tools"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

const (
	sessionid_timeout int64 = 60 * 15          // 15 minutes
	clean_timeout     int64 = 60 * 60 * 24 * 7 // 7 days
)

// CacheInstance is a global variable that is used to access the cache.
var CacheInstance *Cache = NewCache(config.ConfigInstance.Cache, config.ConfigInstance.CachePath)

type Authorization struct {
	App_token          string `db:"app_token"`
	C_token            string `db:"c_token,omitempty"`
	ExpiresAt          int64  `db:"expires_at,omitempty"`
	Vscode_machineid   string `db:"vscode_machineid,omitempty"`
	Vscode_sessionid   string `db:"vscode_sessionid,omitempty"`
	Session_expires_at int64  `db:"session_expires_at,omitempty"`
	Last_touched       int64  `db:"last_touched,omitempty"`
}

// Cache is a struct that contains the cache information.
type Cache struct {
	cache      bool
	cache_path string
	Db         *sqlx.DB
	Data       map[string]Authorization
}

// Create a new Cache instance.
func NewCache(cache bool, cache_path string) *Cache {
	c := &Cache{
		cache:      cache,
		cache_path: cache_path,
	}
	return c
}

// Create table
func (c *Cache) createTable() error {
	if !c.cache {
		return errors.New("cache is disabled")
	}
	if c.Db == nil {
		return errors.New("database is not connected")
	}
	// create table
	_, err := c.Db.Exec(`
		CREATE TABLE IF NOT EXISTS cache(
			app_token TEXT PRIMARY KEY,
			c_token TEXT DEFAULT '',
			expires_at INTEGER DEFAULT 0,
			vscode_machineid TEXT DEFAULT '',
			vscode_sessionid TEXT DEFAULT '',
			session_expires_at INTEGER DEFAULT 0,
			last_touched INTEGER DEFAULT 0
		)
	`)
	return err
}

// Connect to the database or initialize the map
func (c *Cache) connect() {
	if c.cache && c.Db == nil {
		// create cache directory if not exists
		if err := tools.MkdirAllIfNotExists(c.cache_path, os.ModePerm); err != nil {
			log.ZLog.Log.Error().Err(err).Msg("Create cache directory failed, cache_path: " + c.cache_path + ". Please check the configuration file.")
			panic(err)
		}

		// connect to database
		var err error
		c.Db, err = sqlx.Connect("sqlite", c.cache_path)
		if err != nil {
			log.ZLog.Log.Error().Err(err).Msg("Connect to database failed. Please check cache database path, cache_path: " + c.cache_path)
			panic(err)
		}
		// create table if not exists
		// _, err = c.Db.Exec("CREATE TABLE IF NOT EXISTS cache(app_token TEXT PRIMARY KEY, c_token TEXT, expires_at INTEGER, vscode_machineid TEXT, vscode_sessionid TEXT, session_expires_at INTEGER, last_touched INTEGER)") and setdefault
		err = c.createTable()
		if err != nil {
			log.ZLog.Log.Error().Err(err).Msg("Create cache table failed.")
			panic(err)
		}
		// recreate table if not exists
		// check if vscode_machineid exists
		var vscode_machineid string
		err = c.Db.Get(&vscode_machineid, "SELECT vscode_machineid FROM cache LIMIT 1")
		if err != nil && err.Error() != "sql: no rows in result set" {
			// drop table and recreate
			_, err = c.Db.Exec("DROP TABLE cache")
			if err != nil {
				log.ZLog.Log.Error().Err(err).Msg("drop cache table failed")
				panic(err)
			}
			err = c.createTable()
			if err != nil {
				log.ZLog.Log.Error().Err(err).Msg("create cache table failed")
				panic(err)
			}
		}
	} else if !c.cache && c.Data == nil {
		c.Data = make(map[string]Authorization)
	}
}

// get record
func (c *Cache) get(app_token string) (Authorization, bool) {
	c.connect()
	if c.cache {
		var authorization Authorization
		err := c.Db.Get(&authorization, "SELECT * FROM cache WHERE app_token = ?", app_token)
		if err != nil {
			log.ZLog.Log.Warn().Msg("Get authorization from cache failed, app_token: " + app_token)
			return Authorization{}, false
		}
		return authorization, true
	} else {
		if authorization, ok := c.Data[app_token]; ok {
			return authorization, true
		}
		return Authorization{}, false
	}
}

// update authorization to item (ignore zero value)
func (c *Cache) updateItem(item *Authorization, authorization Authorization) {
	if authorization.App_token != "" {
		item.App_token = authorization.App_token
	}
	if authorization.C_token != "" {
		item.C_token = authorization.C_token
	}
	if authorization.ExpiresAt != 0 {
		item.ExpiresAt = authorization.ExpiresAt
	}
	if authorization.Vscode_machineid != "" {
		item.Vscode_machineid = authorization.Vscode_machineid
	}
	if authorization.Vscode_sessionid != "" {
		item.Vscode_sessionid = authorization.Vscode_sessionid
	}
	if authorization.Session_expires_at != 0 {
		item.Session_expires_at = authorization.Session_expires_at
	}
	if authorization.Last_touched != 0 {
		item.Last_touched = authorization.Last_touched
	}
}

// update record
func (c *Cache) modify(app_token string, authorization Authorization) error {
	log.ZLog.Log.Debug().Msg(fmt.Sprint("Modify cache record, app_token: ", app_token, ", authorization: ", authorization))
	c.connect()
	// try to get
	item, ok := c.get(app_token)
	if c.cache {
		if ok {
			c.updateItem(&item, authorization)
			// update
			_, err := c.Db.Exec("UPDATE cache SET c_token = ?, expires_at = ?, vscode_machineid = ?, vscode_sessionid = ?, session_expires_at = ?, last_touched = ? WHERE app_token = ?", item.C_token, item.ExpiresAt, item.Vscode_machineid, item.Vscode_sessionid, item.Session_expires_at, item.Last_touched, app_token)
			if err != nil {
				log.ZLog.Log.Error().Err(err).Msg("update cache failed, app_token: " + app_token)
				return err
			}
		} else {
			// insert
			_, err := c.Db.Exec("INSERT INTO cache VALUES (?, ?, ?, ?, ?, ?, ?)", app_token, authorization.C_token, authorization.ExpiresAt, authorization.Vscode_machineid, authorization.Vscode_sessionid, authorization.Session_expires_at, authorization.Last_touched)
			if err != nil {
				log.ZLog.Log.Error().Err(err).Msg("Insert cache failed, app_token: " + app_token)
				return err
			}
		}
	} else {
		if ok {
			c.updateItem(&item, authorization)
			// update
			c.Data[app_token] = item
		} else {
			// insert
			authorization.App_token = app_token
			c.Data[app_token] = authorization
		}
	}
	return nil
}

// update_session sessionid
func (c *Cache) update_session(app_token string) bool {
	c.connect()
	to_update := Authorization{Last_touched: time.Now().Unix()}
	item, ok := c.get(app_token)
	if !ok {
		return false
	}
	// if session expires, update
	if item.Session_expires_at < time.Now().Unix() {
		to_update.Vscode_sessionid = uuid.NewString() + strconv.FormatInt(time.Now().UnixMilli(), 10)
		to_update.Session_expires_at = time.Now().Unix() + sessionid_timeout
		log.ZLog.Log.Debug().Msg(fmt.Sprintf("Session expires, update sessionid, app_token: %s, sessionid: %s, session_expires_at: %d", app_token, to_update.Vscode_sessionid, to_update.Session_expires_at))
	}
	if item.Vscode_machineid == "" {
		to_update.Vscode_machineid = tools.GenMachineId()
		log.ZLog.Log.Debug().Msg(fmt.Sprintf("Vscode_machineid is empty, generate new one, app_token: %s, vscode_machineid: %s", app_token, to_update.Vscode_machineid))
	}
	// update
	err := c.modify(app_token, to_update)
	return err == nil
}

// clean items that have not been touched for a long time
func (c *Cache) clean() {
	c.connect()
	expirationTime := time.Now().Unix() - clean_timeout

	logExpiredItem := func(app_token string) {
		logMessage := "Clean app_token: %s, the item has not been touched for a long time"
		log.ZLog.Log.Debug().Msg(fmt.Sprintf(logMessage, app_token))
	}

	if c.cache {
		// get all items with last_touched < now - clean_timeout
		var items []Authorization
		err := c.Db.Select(&items, "SELECT * FROM cache WHERE last_touched < ?", time.Now().Unix()-clean_timeout)
		if err == nil {
			// delete
			for _, item := range items {
				c.Delete(item.App_token)
				logExpiredItem(item.App_token)
			}
		}
	} else {
		// get all items with last_touched < now - clean_timeout
		for app_token, item := range c.Data {
			if item.Last_touched < expirationTime {
				c.Delete(app_token)
				logExpiredItem(item.App_token)
			}
		}
	}
}

// Get the Authorization from the cache.
func (c *Cache) Get(app_token string) (Authorization, bool) {
	auth, ok := c.get(app_token)
	if !ok {
		return auth, ok
	}

	log.ZLog.Log.Debug().Msg(fmt.Sprintf("Get authorization from cache success, app_token: %s, will update session expire time", app_token))
	ok = c.update_session(app_token)

	old_auth_return_message := fmt.Sprint("Return the original authorization, app_token: ", app_token, ", authorization: ", auth)
	if !ok {
		log.ZLog.Log.Error().Msg(fmt.Sprintf("Update session failed, %s", old_auth_return_message))
		return auth, ok
	}
	item, ok := c.get(app_token)
	if !ok {
		log.ZLog.Log.Error().Msgf("Get authorization from cache failed, %s", old_auth_return_message)
		return auth, ok
	}
	return item, ok
}

// Set the Authorization in the cache.
func (c *Cache) Set(app_token string, authorization Authorization) bool {
	c.connect()
	now := time.Now().Unix()
	item, ok := c.get(app_token)
	if ok { // if exists, update
		item.C_token = authorization.C_token
		item.ExpiresAt = authorization.ExpiresAt
		item.Last_touched = now
	} else {
		item = Authorization{
			App_token:        app_token,
			C_token:          authorization.C_token,
			ExpiresAt:        authorization.ExpiresAt,
			Vscode_machineid: tools.GenMachineId(),
			Last_touched:     now,
		}
	}
	err := c.modify(app_token, item)
	if err != nil {
		return false
	}
	return c.update_session(app_token)
}

// Delete the Authorization from the cache.
func (c *Cache) Delete(app_token string) error {
	c.connect()
	if c.cache {
		_, err := c.Db.Exec("DELETE FROM cache WHERE app_token = ?", app_token)
		if err != nil {
			log.ZLog.Log.Error().Err(err).Msg("Delete cache item failed, app_token: " + app_token)
			return err
		}
		return nil
	} else {
		delete(c.Data, app_token)
		return nil
	}
}

// Close the database connection.
func (c *Cache) Close() {
	if c.cache && c.Db != nil {
		c.clean()
		c.Db.Close()
		c.Db = nil
	} else if !c.cache && c.Data != nil {
		c.clean()
		c.Data = nil
	}
}
