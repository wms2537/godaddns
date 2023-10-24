package storage

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

var db *badger.DB

func init() {
	// Open the database with appropriate configuration.
	mydb, err := badger.Open(badger.DefaultOptions("./badger_data"))
	if err != nil {
		panic(err)
	}
	db = mydb
}

func Close() {
	db.Close()
}

type WhitelistEntry struct {
	Nodes []string
}

func AddUserToWhitelist(username, nodeID string) error {
	var whitelistEntry WhitelistEntry

	// Fetch the current entry or create a new one.
	err := db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(username))
		if err == badger.ErrKeyNotFound {
			whitelistEntry = WhitelistEntry{}
		} else if err != nil {
			return err
		} else {
			valCopy, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(valCopy, &whitelistEntry); err != nil {
				return err
			}
		}

		// Add the node ID to the entry.
		whitelistEntry.Nodes = append(whitelistEntry.Nodes, nodeID)

		// Serialize the updated entry.
		entryBytes, err := json.Marshal(whitelistEntry)
		if err != nil {
			return err
		}

		// Store the updated entry back in the database.
		return txn.Set([]byte(username), entryBytes)
	})

	if err != nil {
		return err
	}

	fmt.Printf("Node ID %s added to the whitelist for user %s.\n", nodeID, username)
	return nil
}

func IsUserNodeInWhitelist(username, nodeID string) (bool, error) {
	var whitelistEntry WhitelistEntry

	isWhitelist := false

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(username))
		if err == badger.ErrKeyNotFound {
			return nil // User not found in the whitelist.
		} else if err != nil {
			return err
		}

		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(valCopy, &whitelistEntry); err != nil {
			return err
		}

		// Check if the node ID is in the user's whitelist.
		for _, id := range whitelistEntry.Nodes {
			if id == nodeID {
				isWhitelist = true
				return nil // Node ID found in the whitelist.
			}
		}

		return nil // Node ID not found in the user's whitelist.
	})

	if err != nil {
		return false, err
	}

	return isWhitelist, nil
}
