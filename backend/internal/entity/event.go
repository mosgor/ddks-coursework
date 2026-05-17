package entity

import "time"

// @Description Описание события
type Event struct {
	Id          int       `json:"id,omitempty"`
	Title       string    `json:"title"`
	Date        time.Time `json:"date"`
	Description string    `json:"description,omitempty"`
	Image       string    `json:"image,omitempty"`
	Price       int       `json:"price"`
}
