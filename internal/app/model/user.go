package model

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"golang.org/x/crypto/bcrypt"
)

//User ...
type User struct {
	ID                int    `json:"id"`
	FirstName         string `json:"sendto"`
	LastName          string `json:"text"`
	Username          string `json:"username"`
	Password          string `json:"password,omitempty"`
	EncruptedPassword string `json:"-"`
	IP                string `json:"-"`
	Avatar            string `json:"avatar"`
	Gender            string `json:"gender"`
	Country           string `json:"country"`
	City              string `json:"city"`
	Bio               string `json:"bio"`
	BirthdayDate      int    `json:"birthday_date"`
	Me                bool   `json:"me"`
	Status            string `json:"status"`
	Online            int    `json:"online"`
}

//SettingsMain ...
type SettingsMain struct {
	ID           int    `json:"id"`
	BirthdayDate int    `json:"birthday_date"`
	Gender       string `json:"gender"`
	Country      string `json:"country"`
	City         string `json:"city"`
	Bio          string `json:"bio"`
}

//Validate ...
func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Username, validation.Required, validation.Length(1, 400)),
		validation.Field(&u.Password, validation.Required),
	)
}

//BeforeCreate ...
func (u *User) BeforeCreate() error {
	if len(u.Password) > 0 {
		enc, err := encryptString(u.Password)
		if err != nil {
			return err
		}

		u.EncruptedPassword = enc
		//u.Sanitize()

		u.Username = strings.TrimSpace(u.Username)
	}
	return nil
}

//Sanitize ...
func (u *User) Sanitize() {
	u.Password = ""
}

//ComparePassword ...
func (u *User) ComparePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.EncruptedPassword), []byte(password)) == nil
}

func encryptString(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)

	if err != nil {
		return "", err
	}

	return string(b), nil
}
