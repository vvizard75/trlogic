package model

type Result struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Error    bool   `json:"error"`
}
