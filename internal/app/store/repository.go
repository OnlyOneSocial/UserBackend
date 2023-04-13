package store

import "github.com/katelinlis/UserBackend/internal/app/model"

// FriendsRepository ...
type FriendsRepository interface {
	SendRequest(*model.Friends) error
	Accept(*model.Friends) error
	GetCount(userid int) (int, error)
	GetCountRequests(userid int) (int, error)
	GetArrayFriends(userid int) ([]int, error)
	Get(userid int) ([]model.Friends, error)
	GetStatusFriend(*model.Friends) error
	GetAllRequests(user int) ([]model.Friends, error)
	GetRequests(userid int) ([]model.Friends, error)
	GetAllFriends(user int) ([]model.Friends, error)
	GetAllSubscribes(user int) ([]model.Friends, error)
}

// UserRepository ...
type UserRepository interface {
	Create(*model.User) error
	SetAvatar(userid int, avatar string) error
	SetIP(userid int, ip string) error
	AddToBanList(user int, whoToBanned int) (bool, error)
	RemoveFromBanList(user int, WhoUnban int) (bool, error)
	ChangeStatus(userid int, status string) error
	ChangeSettingsMain(*model.SettingsMain) error
	UpdateOnline(userid int) error
	ChangePassword(user *model.User) error
	Update(*model.User) error
	Find(userid int) (model.User, error)
	FindByUsername(username string) (model.User, error)
	FindByUsernameLike(username string) ([]model.User, error)
	FindByUsernameAndPassword(u *model.User) error
	GetCount() (int, error)
	Get() ([]model.User, error)
	GetAllUsersWithPassword() (users []model.User, err error)
	UpdateUserConnectToKeyCloak(userid int, keycloak string) error
}
