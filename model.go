package main

type Account string

type Tx struct {
	From  Account `json:"from"`
	To    Account `json:"to"`
	Value float32 `json:"amount"`
}

type Genesis struct {
	Balances map[Account]float32 `json:"balances"`
}
