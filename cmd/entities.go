package main

type User struct {
	Id      int     `db:"id"`
	UserId  int     `json:"id" db:"user_id"`
	Balance float32 `json:"balance" db:"balance"`
}

type CurrencyList struct {
	Valute map[string]map[string]interface{} `json:"Valute"`
}
