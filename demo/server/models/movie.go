package models

type Movie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Released string `json:"released"`
}
