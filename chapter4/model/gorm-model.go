package model

type GormUser struct {
	ID       uint   `json:"id"`
	Phone    string `json:"phone"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
