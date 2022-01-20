// Module florist helps creating non idempotent, one-file-contains-everything
// installers/provisioners.
package florist

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"time"
)

// A flower is a composable unit that can be installed.
type Flower interface {
	fmt.Stringer
	Description() string
	Install() error
}

const (
	WorkDir  = "/tmp/florist.work"
	DataPath = WorkDir + "/florist.json"

	CacheValidityDefault = 24 * time.Hour
	EmbedDir             = "files"
)

func Init() (*user.User, error) {
	log := Log.Named("Init")

	userSelf, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("florist.Init: %s", err)
	}
	if err := Mkdir(WorkDir, userSelf, 0755); err != nil {
		return userSelf, fmt.Errorf("florist.Init: %s", err)
	}
	log.Info("success")
	return userSelf, nil
}

type Record struct {
	Flowers map[string]time.Time `json:"flowers"`
}

//
func ReadRecord() (Record, error) {
	log := Log.Named("ReadRecord")

	_, err := os.Stat(DataPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// if the file doesn't exist, it is OK; return an empty Record
			return Record{Flowers: make(map[string]time.Time)}, nil
		}
		return Record{}, err
	}
	fi, err := os.Open(DataPath)
	if err != nil {
		return Record{}, err
	}
	defer fi.Close()
	var record Record
	dec := json.NewDecoder(fi)
	if err := dec.Decode(&record); err != nil {
		return Record{}, err
	}
	log.Info("Read Record", "record", record)
	return record, nil
}

func WriteRecord(name string) error {
	log := Log.Named("WriteRecord")

	record, err := ReadRecord()
	if err != nil {
		return err
	}

	log.Info("Save Record", "name", name)
	record.Flowers[name] = time.Now()
	fi, err := os.Create(DataPath)
	if err != nil {
		return err
	}
	defer fi.Close()
	enc := json.NewEncoder(fi)
	if err := enc.Encode(&record); err != nil {
		return err
	}

	log.Info("Customize motd")
	motd := "System provisioned by 🌼 florist 🌺\n"
	return os.WriteFile("/etc/motd", []byte(motd), 0644)
}
