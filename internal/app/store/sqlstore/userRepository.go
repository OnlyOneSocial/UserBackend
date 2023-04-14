package sqlstore

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/katelinlis/UserBackend/internal/app/model"
)

// UserRepository ...
type UserRepository struct {
	store *Store
}

// Create ...
func (r *UserRepository) Create(u *model.User) error {

	err := u.Validate()

	if err != nil {
		return err
	}

	err = u.BeforeCreate()

	if err != nil {
		return err
	}

	err = r.store.db.QueryRow("INSERT INTO users(first_name,last_name,username,password,ip) VALUES($1,$2, $3,$4,$5) RETURNING id",
		u.FirstName,
		u.LastName,
		u.Username,
		u.EncruptedPassword,
		u.IP,
	).Scan(&u.ID)

	return err
}

// AddToBanList ...
func (r *UserRepository) AddToBanList(user int, whoToBanned int) (bool, error) {
	err := r.store.db.QueryRow("UPDATE users SET banlist = array_append(banlist,$1) WHERE id=$2 RETURNING id", whoToBanned, user)

	var id int
	if err := err.Scan(&id); err != nil {
		return false, err
	}

	return true, nil
}

// RemoveToBanList ...
func (r *UserRepository) RemoveFromBanList(user int, WhoUnban int) (bool, error) {
	err := r.store.db.QueryRow("UPDATE users SET banlist = array_remove(banlist,$1) WHERE id=$2 RETURNING id", WhoUnban, user)

	var id int
	if err := err.Scan(&id); err != nil {
		return false, err
	}

	return true, nil
}

// SetAvatar ...
func (r *UserRepository) SetAvatar(userid int, avatar string) error {
	var id int
	err := r.store.db.QueryRow("UPDATE users SET avatar = $1 where id=$2 RETURNING id",
		avatar,
		userid,
	).Scan(&id)

	if id == userid {
		fmt.Println(id == userid)
	}

	return err
}

// SetIP ...
func (r *UserRepository) SetIP(userid int, ip string) error {
	var id int
	err := r.store.db.QueryRow("UPDATE users SET avatar = $1 where id=$2 RETURNING id",
		ip,
		userid,
	).Scan(&id)

	if id == userid {
		fmt.Println(id == userid)
	}

	return err
}

// Update ...
func (r *UserRepository) Update(u *model.User) error {
	err := r.store.db.QueryRow("UPDATE users SET first_name = $1,last_name = $2,username = $3, where id=$4 RETURNING id",
		u.FirstName,
		u.LastName,
		u.Username,
		u.ID,
	).Scan(&u.ID)

	return err
}

// ChangeStatus ...
func (r *UserRepository) ChangeStatus(userid int, status string) error {
	var id int
	err := r.store.db.QueryRow("UPDATE users SET status = $1 where id=$2 RETURNING id",
		status,
		userid,
	).Scan(&id)
	return err
}

// ChangePassword ...
func (r *UserRepository) ChangePassword(user *model.User) error {
	var id int

	user.BeforeCreate()

	err := r.store.db.QueryRow("UPDATE users SET password = $1 where id=$2 RETURNING id",
		user.EncruptedPassword,
		user.ID,
	).Scan(&id)
	return err
}

// UpdateOnline ...
func (r *UserRepository) UpdateOnline(userid int) error {
	var id int
	err := r.store.db.QueryRow("UPDATE users SET online = $1 where id=$2 RETURNING id",
		time.Now().Unix(),
		userid,
	).Scan(&id)
	return err
}

//UpdateUserConnectToKeyCloak ...
func (r *UserRepository) UpdateUserConnectToKeyCloak(userid int, keycloak string) error {
	var id int
	err := r.store.db.QueryRow("UPDATE users SET SSOID = $1 where id=$2 RETURNING id",
		keycloak,
		userid,
	).Scan(&id)
	return err
}

// Find ...
func (r *UserRepository) Find(userid int) (model.User, error) {
	var user model.User
	var avatar sql.NullString
	var status sql.NullString
	var gender sql.NullString
	var country sql.NullString
	var city sql.NullString
	var bio sql.NullString
	var birthday sql.NullInt64
	var online sql.NullInt64
	err := r.store.db.QueryRow("SELECT username,id,first_name,last_name,avatar,status,gender,country,city,bio,birthday,online from users where id = $1",
		userid,
	).Scan(&user.Username, &user.ID, &user.FirstName, &user.LastName, &avatar, &status, &gender, &country, &city, &bio, &birthday, &online)
	user.Avatar = avatar.String
	user.Status = status.String
	user.Gender = gender.String
	user.Country = country.String
	user.City = city.String
	user.Bio = bio.String
	user.Online = int(online.Int64)
	user.BirthdayDate = int(birthday.Int64)

	return user, err
}

// Get ...
func (r *UserRepository) Get() ([]model.User, error) {
	var users []model.User

	rows, err := r.store.db.Query("SELECT username,id,avatar from users order by id")
	if err != nil {
		return users, err
	}

	for rows.Next() {
		user := model.User{}
		var avatar sql.NullString
		err := rows.Scan(&user.Username, &user.ID, &avatar)
		user.Avatar = avatar.String
		if err != nil {
			return users, err
		}

		users = append(users, user)

	}
	return users, err
}

// FindByUsername ...
func (r *UserRepository) FindByUsername(username string) (model.User, error) {
	var user model.User
	var avatar sql.NullString
	err := r.store.db.QueryRow("SELECT username,id,password,avatar FROM users WHERE LOWER(username) LIKE LOWER($1)",
		username,
	).Scan(&user.Username, &user.ID, &user.EncruptedPassword, &avatar)
	user.Avatar = avatar.String

	return user, err
}

//FindByUsernameLike ...
func (r *UserRepository) FindByUsernameLike(username string) (users []model.User, err error) {
	query := `SELECT username,id,avatar FROM users WHERE lower(username) LIKE '%' || lower($1) || '%' order by id`

	rows, err := r.store.db.Query(query, username)
	if err != nil {
		return users, err
	}

	for rows.Next() {
		user := model.User{}
		var avatar sql.NullString
		err := rows.Scan(&user.Username, &user.ID, &avatar)
		user.Avatar = avatar.String
		if err != nil {
			return users, err
		}

		users = append(users, user)

	}
	return users, err
}

// ChangeSettingsMain ...
func (r *UserRepository) ChangeSettingsMain(settings *model.SettingsMain) error {

	err := r.store.db.QueryRow("UPDATE users SET country = $2,city = $3,gender = $4,bio = $5, birthday = $6 where id=$1 RETURNING id",
		settings.ID,
		settings.Country,
		settings.City,
		settings.Gender,
		settings.Bio,
		settings.BirthdayDate,
	).Scan(&settings.ID)
	return err
}

// GetAllUsersWithPassword ...
func (r *UserRepository) GetAllUsersWithPassword() (users []model.User, err error) {

	rows, err := r.store.db.Query("SELECT username,id,password from users where ssid is NULL")
	if err != nil {
		return users, err
	}

	for rows.Next() {
		user := model.User{}
		err = rows.Scan(&user.Username, &user.ID, &user.EncruptedPassword)
		if err != nil {
			return users, err
		}

		users = append(users, user)
	}

	return users, err
}

// FindByUsernameAndPassword ...
func (r *UserRepository) FindByUsernameAndPassword(u *model.User) error {

	err := r.store.db.QueryRow("SELECT username,id,first_name,last_name,avatar,user_location where username = $1 and password = $2",
		u.Username,
		u.EncruptedPassword,
	).Scan(&u.Username, &u.ID, &u.FirstName, &u.LastName, &u.Avatar, &u.City)

	return err
}

// GetCount ...
func (r *UserRepository) GetCount() (int, error) {
	var count int
	err := r.store.db.QueryRow("SELECT Count (*) from users").Scan(&count)

	return count, err
}
