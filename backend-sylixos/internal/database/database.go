package database

import (
	"go-ser/internal/config"
)

var EdgeDB *BoltDB

func Start() error {
	var err error
	EdgeDB, err = newBoltDB(config.AppConfig.Database.EdgePath)
	if err != nil {
		return err
	}
	return nil
}

func Stop() {
	if EdgeDB != nil {
		EdgeDB.Close()
	}
}
