package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

func NewStart() bool {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Permission denied", err)
	}

	_, err = os.Stat(filepath.Join(usr.HomeDir, "state.json"))
	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(filepath.Join(usr.HomeDir, "state.json"))
			if err != nil {
				log.Fatal("state.json creation error", err)
			}
		}
	}
	_, err = os.Stat(filepath.Join(usr.HomeDir, "tx.db"))
	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(filepath.Join(usr.HomeDir, "tx.db"))
			if err != nil {
				log.Fatal("state.json creation error", err)
			}
			return true
		}
	}
	return false
}

func CreateNewAccount() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Permission denied", err)
	}

	file, err := os.OpenFile(filepath.Join(usr.HomeDir, "state.json"), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		log.Fatal("state.json open error", err)
	}

	gen := Genesis{Balances: map[Account]float32{}}
	gen.Balances = make(map[Account]float32)
	for i := 0; i < 10; i++ {
		privateKeyECDSA, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
		key := &keystore.Key{
			Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
			PrivateKey: privateKeyECDSA,
		}
		addres := hex.EncodeToString(key.Address[:])
		gen.Balances[NewAccount(addres)] = 100

		rawDataOut, err := json.MarshalIndent(&Genesis{gen.Balances}, "", "  ")
		if err != nil {
			log.Fatal("JSON marshaling failed:", err)
		}

		err = ioutil.WriteFile(filepath.Join(usr.HomeDir, "state.json"), rawDataOut, 0)
		if err != nil {
			log.Fatal("Cannot write updated settings file:", err)
		}
	}

	file.Close()
}

func NewAccount(value string) Account {
	return Account(value)
}
