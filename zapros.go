package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

func getBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "/api/wallet/{address}/balance")
	address := mux.Vars(r)
	ac := NewAccount(address["address"])

	usr, err := user.Current()
	if err != nil {
		log.Fatal("Permission denied", err)
	}

	gen, err := loadGenesis(filepath.Join(usr.HomeDir, "state.json"))
	if err != nil {
		fmt.Fprintf(w, "read error: %s", err)
		return
	}

	if count, ok := gen.Balances[ac]; ok {
		json.NewEncoder(w).Encode(count)
		return
	}
	json.NewEncoder(w).Encode(&Genesis{})
}

func getLast(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil || id < 1 {
		fmt.Fprintf(w, "wrong counter %d. counter > 0", id)
		return
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal("Permission denied", err)
	}

	fil_lin, err := os.OpenFile(filepath.Join(usr.HomeDir, "tx.db"), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		fmt.Fprintf(w, "tx.db open error: %s", err)
		return
	}
	line, err := lineCounter(fil_lin)
	if err != nil {
		fmt.Fprintf(w, "tx.db read error: %s", err)
		return
	}
	fil_lin.Close()

	f, err := os.OpenFile(filepath.Join(usr.HomeDir, "tx.db"), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		fmt.Fprintf(w, "tx.db open error: %s", err)
		return
	}
	scanner := bufio.NewScanner(f)
	n := line - id
	lin := 0
	for scanner.Scan() {
		lin++
		if lin > n {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(w, "tx.db red error: %s", err)
				return
			}
			var tx Tx
			err = json.Unmarshal(scanner.Bytes(), &tx)
			if err != nil {
				fmt.Fprintf(w, "tx.db red error: %s", err)
				return
			}
			json.NewEncoder(w).Encode(&tx)
		}
	}
}

func lineCounter(r io.Reader) (int, error) {

	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

func postSend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "/api/send")

	var tx Tx
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "specify in the format: {\"from\":\"adress\",\"to\":\"adress\",\"amount\":transfer amount}")
		return
	} else {
		err = json.Unmarshal(body, &tx)
		if err != nil {
			fmt.Fprintf(w, "specify in the format: {\"from\":\"adress\",\"to\":\"adress\",\"amount\":transfer amount}")
			return
		}
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal("Permission denied", err)
	}

	gen, err := loadGenesis(filepath.Join(usr.HomeDir, "state.json"))
	if err != nil {
		fmt.Fprintf(w, "read error: %s", err)
		return
	}

	balances := make(map[Account]float32)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	if tx.Value < 0 {
		fmt.Fprintf(w, "transfer amount > 0")
		return
	}
	_, ok := gen.Balances[tx.From]
	if !ok {
		fmt.Fprintf(w, "wallet with specified address does not exist")
		return
	}
	_, ok = gen.Balances[tx.To]
	if !ok {
		fmt.Fprintf(w, "wallet with specified address does not exist")
		return
	}

	if gen.Balances[tx.From]-tx.Value < 0 {
		fmt.Fprintf(w, "insufficient funds to transfer: balance < amount")
		return
	}

	gen.Balances[tx.From] -= tx.Value
	gen.Balances[tx.To] += tx.Value

	rawDataOut, err := json.MarshalIndent(&Genesis{gen.Balances}, "", "  ")
	if err != nil {
		log.Fatal("JSON marshaling failed:", err)
	}

	err = ioutil.WriteFile(filepath.Join(usr.HomeDir, "state.json"), rawDataOut, 0)
	if err != nil {
		log.Fatal("Cannot write updated settings file:", err)
	}

	f, err := os.OpenFile(filepath.Join(usr.HomeDir, "tx.db"), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		log.Fatal("tx.db open error", err)
	}
	_, err = f.WriteString(string(body))
	if err != nil {
		log.Fatal("Cannot write updated settings file:", err)
	}
}

func loadGenesis(path string) (Genesis, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Genesis{}, err
	}

	var loadedGenesis Genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return Genesis{}, err
	}

	return loadedGenesis, nil
}
