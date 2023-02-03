package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

//Friends ...
type Friends struct {
	User1     int   `json:"userid"`
	User2     int   `json:"user2"`
	Timestamp int64 `json:"timestamp"`
	Status    int   `json:"status"`
	ForMe     bool  `json:"forme"`
	User      User  `json:"user"`
}

//Validate ...
func (w *Friends) Validate() error {
	return validation.ValidateStruct(
		w,
		validation.Field(&w.User1, validation.Required, validation.Length(1, 400)),
		validation.Field(&w.User2, validation.Required, validation.Length(1, 400)),
	)
}
