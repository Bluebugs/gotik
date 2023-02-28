package main

import (
	"crypto/rand"
	"fmt"
	"log"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"go.etcd.io/bbolt"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"
)

func (a *appData) openDB() error {
	dbURI, err := storage.Child(a.app.Storage().RootURI(), "network.boltdb")
	if err != nil {
		return err
	}

	a.db, err = bbolt.Open(dbURI.Path(), 0600, nil)
	if err != nil {
		return err
	}

	a.restoreCurrentView()

	return nil
}

func (a *appData) saveRouter(r *router, password string) error {
	return a.db.Update(func(tx *bbolt.Tx) error {
		routers, err := tx.CreateBucketIfNotExists([]byte("routers"))
		if err != nil {
			return err
		}

		host, err := routers.CreateBucketIfNotExists([]byte(r.host))
		if err != nil {
			return err
		}

		err = host.Put([]byte("user"), a.key.Seal([]byte(r.user)))
		if err != nil {
			return err
		}

		err = host.Put([]byte("password"), a.key.Seal([]byte(password)))
		if err != nil {
			return err
		}

		return nil
	})
}

func (a *appData) saveCurrentView() error {
	return a.db.Update(func(tx *bbolt.Tx) error {
		settings, err := tx.CreateBucketIfNotExists([]byte("settings"))
		if err != nil {
			return err
		}

		return settings.Put([]byte("currentView"), []byte(a.currentView))
	})
}

func (a *appData) restoreCurrentView() error {
	return a.db.View(func(tx *bbolt.Tx) error {
		settings := tx.Bucket([]byte("settings"))
		if settings == nil {
			return nil
		}

		currentView := settings.Get([]byte("currentView"))
		if currentView == nil {
			return nil
		}

		a.currentView = string(currentView)
		return nil
	})
}

func (a *appData) loadRouters(sel *widget.Select) error {
	return a.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("routers"))
		if b == nil {
			log.Println("no routers in network.boltdb")
			return nil
		}

		err := b.ForEach(func(k, v []byte) error {
			host := string(k)
			b := b.Bucket(k)
			if b == nil {
				log.Println("incorrect host", host, "in network.boltdb, skipping")
				return nil
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

			r, err := routerView(host, string(user), string(password))
			if err != nil {
				dialog.ShowError(err, a.win)
				return nil
			}

			a.routers[host] = r
			sel.Options = append(sel.Options, r.host)
			return nil
		})
		if err != nil {
			return err
		}
		sel.Refresh()
		return nil
	})
}

func (a *appData) salt() (s []byte) {
	a.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("settings"))
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
		b, err := tx.CreateBucketIfNotExists([]byte("settings"))
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
