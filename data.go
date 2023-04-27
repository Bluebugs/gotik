package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"go.etcd.io/bbolt"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"
)

var settingBucketName = []byte("settings")
var routersBucketName = []byte("routers")

func (a *appData) openDB() (string, error) {
	dbURI, err := storage.Child(a.app.Storage().RootURI(), "network.boltdb")
	if err != nil {
		return "", err
	}

	a.db, err = bbolt.Open(dbURI.Path(), 0600, nil)
	if err != nil {
		return "", err
	}

	return a.restoreCurrentView()
}

func saveHost(tx *bbolt.Tx, key *secretKey, host string, ssl bool, user string, password string) error {
	routers, err := tx.CreateBucketIfNotExists(routersBucketName)
	if err != nil {
		return err
	}

	hostBucket, err := routers.CreateBucketIfNotExists([]byte(host))
	if err != nil {
		return err
	}

	var sslBytes []byte
	if ssl {
		sslBytes = key.Seal([]byte("true"))
	} else {
		sslBytes = key.Seal([]byte("false"))
	}
	err = hostBucket.Put([]byte("ssl"), sslBytes)
	if err != nil {
		return err
	}

	err = hostBucket.Put([]byte("user"), key.Seal([]byte(user)))
	if err != nil {
		return err
	}

	err = hostBucket.Put([]byte("password"), key.Seal([]byte(password)))
	if err != nil {
		return err
	}

	return nil
}

func (a *appData) saveRouter(r *router, password string) error {
	return a.db.Update(func(tx *bbolt.Tx) error {
		return saveHost(tx, a.key, r.host, r.ssl, r.user, password)
	})
}

func deleteHost(tx *bbolt.Tx, host string) error {
	routers := tx.Bucket(routersBucketName)
	if routers == nil {
		return nil
	}

	return routers.DeleteBucket([]byte(host))
}

func (a *appData) deleteRouter(r *router) error {
	return a.db.Update(func(tx *bbolt.Tx) error {
		return deleteHost(tx, r.host)
	})
}

func (a *appData) saveCurrentView() error {
	return a.db.Update(func(tx *bbolt.Tx) error {
		settings, err := tx.CreateBucketIfNotExists(settingBucketName)
		if err != nil {
			return err
		}

		if a.useTailScale {
			settings.Put([]byte("useTailScale"), []byte("true"))
		} else {
			settings.Put([]byte("useTailScale"), []byte("false"))
		}

		if a.current != nil {
			settings.Put([]byte("currentHost"), []byte(a.current.host))
		}

		settings.Put([]byte("currentTab"), []byte(a.currentTab))

		return settings.Put([]byte("currentView"), []byte(a.currentView))
	})
}

func (a *appData) restoreCurrentView() (string, error) {
	var r string
	return r, a.db.View(func(tx *bbolt.Tx) error {
		settings := tx.Bucket(settingBucketName)
		if settings == nil {
			return nil
		}

		useTailScale := settings.Get([]byte("useTailScale"))
		if useTailScale != nil {
			a.useTailScale = string(useTailScale) == "true"
		}

		currentHost := settings.Get([]byte("currentHost"))
		if currentHost != nil {
			r = string(currentHost)
		}

		currentTab := settings.Get([]byte("currentTab"))
		if currentTab != nil {
			a.currentTab = string(currentTab)
		}

		currentView := settings.Get([]byte("currentView"))
		if currentView != nil {
			a.currentView = string(currentView)
		}

		return nil
	})
}

func (a *appData) loadRouters(sel *widget.Select) error {
	resave := []*router{}

	err := a.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(routersBucketName)
		if b == nil {
			log.Println("no routers in network.boltdb")
			return nil
		}

		err := b.ForEach(func(k, v []byte) error {
			host := string(k)
			needResave := false

			b := b.Bucket(k)
			if b == nil {
				log.Println("incorrect host", host, "in network.boltdb, skipping")
				return nil
			}

			ssl := false
			cipherSSL := b.Get([]byte("ssl"))
			if cipherSSL == nil {
				host = strings.TrimSuffix(host, ":8728")
				needResave = true
			} else {
				clearSSL, ok := a.key.Unseal(cipherSSL)
				if !ok {
					return fmt.Errorf("invalid ssl, network.boltdb is corrupted")
				}

				ssl = string(clearSSL) == "true"
			}

			cipherUser := b.Get([]byte("user"))
			if cipherUser == nil {
				log.Println("incorrect user for host", host, "in network.boltdb, skipping")
				return nil
			}

			cipherPassword := b.Get([]byte("password"))
			if cipherPassword == nil {
				log.Println("incorrect password for host", host, "in network.boltdb, skipping")
				return nil
			}

			user, ok := a.key.Unseal(cipherUser)
			if !ok {
				return fmt.Errorf("invalid user, network.boltdb is corrupted")
			}

			password, ok := a.key.Unseal(cipherPassword)
			if !ok {
				return fmt.Errorf("invalid password, network.boltdb is corrupted")
			}

			r := a.routerView(host, ssl, string(user), string(password))
			if r.err != nil {
				dialog.ShowError(r.err, a.win)
			}

			if needResave {
				resave = append(resave, r)
			}
			a.routers[host] = r
			sel.Options = append(sel.Options, r.host)
			return nil
		})
		if err != nil {
			return err
		}
		sel.Refresh()
		a.updateSystray(sel)
		return nil
	})
	if err != nil {
		return err
	}

	if len(resave) == 0 {
		return nil
	}

	return a.db.Update(func(tx *bbolt.Tx) error {
		for _, r := range resave {
			saveHost(tx, a.key, r.host, r.ssl, r.user, r.password)

			deleteHost(tx, r.host+":8728")
		}
		return nil
	})
}

func (a *appData) salt() (s []byte) {
	a.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(settingBucketName)
		if b == nil {
			return nil
		}

		s = b.Get([]byte("salt"))
		if s == nil {
			return nil
		}

		if len(s) != 32 {
			s = nil
			return fmt.Errorf("invalid salt, network.boltdb is corrupted")
		}

		return nil
	})
	return
}

func (a *appData) createKey(password string) error {
	key := [32]byte{}
	_, err := rand.Read(key[:])
	if err != nil {
		panic(err)
	}

	err = a.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(settingBucketName)
		if err != nil {
			return err
		}

		return b.Put([]byte("salt"), key[:])
	})
	if err != nil {
		return err
	}

	a.key = newSecretKey(password, key[:])
	return nil
}

func (a *appData) unlockKey(password string) error {
	salt := a.salt()
	if salt == nil {
		return fmt.Errorf("no salt found, network.boltdb is corrupted")
	}

	a.key = newSecretKey(password, salt)
	return nil
}

func hashPassword(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
}

type secretKey [32]byte

func newSecretKey(password string, salt []byte) *secretKey {
	r := secretKey(hashPassword([]byte(password), salt))
	return &r
}

func (s *secretKey) Seal(plaintext []byte) []byte {
	nonce := make([]byte, 24)
	_, err := rand.Read(nonce)
	if err != nil {
		panic(err)
	}
	return secretbox.Seal(nonce, plaintext, (*[24]byte)(nonce), (*[32]byte)(s))
}

func (s *secretKey) Unseal(ciphertext []byte) ([]byte, bool) {
	nonce := ciphertext[:24]
	plaintext, ok := secretbox.Open(nil, ciphertext[24:], (*[24]byte)(nonce), (*[32]byte)(s))
	return plaintext, ok
}
