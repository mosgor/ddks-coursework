package entity

// @Description Описание пользователя
type User struct {
	Id       int    `json:"id,omitempty"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
