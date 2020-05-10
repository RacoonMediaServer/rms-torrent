package accounts

import (
	"racoondev.tk/gitea/racoon/rms-shared/pkg/db"
	"racoondev.tk/gitea/racoon/rms-shared/pkg/settings"
	"sync"
)

var cache struct {
	mutex    sync.Mutex
	accounts map[string]settings.TorrentAccount
}

func Load(database *db.Database) error {
	accounts, err := database.LoadAccounts()

	if err != nil {
		return nil
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.accounts = make(map[string]settings.TorrentAccount)
	for _, account := range accounts {
		cache.accounts[account.ID] = account
	}
	return nil
}

func Get(trackerID string) (login string, password string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if cache.accounts == nil {
		return
	}

	account, ok := cache.accounts[trackerID]
	if ok {
		login = account.User
		password = account.Password
	}

	return
}
