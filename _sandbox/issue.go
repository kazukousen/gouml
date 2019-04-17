package main

type foo struct {
	id  int `json:"id"`
	bar struct {
		id  int `json:"bar_id"`
		baz struct {
			id int `json:"bar_baz_id"`
		}
	} `json:"bar"`
}
