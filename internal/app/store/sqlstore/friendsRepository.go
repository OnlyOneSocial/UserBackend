package sqlstore

import (
	"errors"
	"time"

	"github.com/katelinlis/UserBackend/internal/app/model"
)

//FriendsRepository ...
type FriendsRepository struct {
	store *Store
}

//SendRequest ...
func (r *FriendsRepository) SendRequest(f *model.Friends) error {
	if f.User1 == f.User2 {
		return errors.New("friend request you can't send to yourself")
	}

	err := r.GetStatusFriend(f)

	if f.Status == 0 && f.ForMe {
		r.Accept(f)
	}

	if f.Status == 0 && !f.ForMe {
		return errors.New("You can`t send request again")
	}

	err = r.store.db.QueryRow("INSERT INTO friends(user1,user2,status,timestamp) VALUES($1,$2,0,$3) RETURNING status",
		f.User1,
		f.User2,
		time.Now().Unix(),
	).Scan(&f.Status)

	return err
}

//Accept ...
func (r *FriendsRepository) Accept(f *model.Friends) error {
	if f.User1 == f.User2 {
		return errors.New("friend request you can't send to yourself")
	}
	err := r.store.db.QueryRow("UPDATE friends SET status = 1 where user2=$1 and status=0 and user1=$2 RETURNING status",
		f.User1,
		f.User2,
	).Scan(&f.Status)

	return err
}

//GetCount ...
func (r *FriendsRepository) GetCount(userid int) (int, error) {
	var count int
	err := r.store.db.QueryRow("SELECT COUNT(*) from friends  where (user2=$1 OR user1=$1) and status=1",
		userid,
	).Scan(&count)

	return count, err
}

//GetCountRequests ...
func (r *FriendsRepository) GetCountRequests(userid int) (int, error) {
	var count int
	err := r.store.db.QueryRow("SELECT COUNT(*) from friends  where user2=$1 and status=0",
		userid,
	).Scan(&count)

	return count, err
}

//GetArrayFriends ...
func (r *FriendsRepository) GetArrayFriends(userid int) ([]int, error) {
	var ids []int

	friends, err := r.Get(userid)
	if err != nil {
		return ids, err
	}

	for _, friend := range friends {
		var id int

		if friend.User1 == userid {
			id = friend.User2
		}

		if friend.User2 == userid {
			id = friend.User1
		}

		ids = append(ids, id)
	}

	follows, err := r.GetAllRequests(userid)
	if err != nil {
		return ids, err
	}

	for _, follow := range follows {
		var id int
		matchOnIDS := false

		if follow.User1 == userid {
			id = follow.User2
		}

		if follow.User2 == userid {
			id = follow.User1
		}

		for _, idInIDS := range ids {
			if idInIDS != id {
				matchOnIDS = true
			}
		}
		if matchOnIDS == false {
			ids = append(ids, id)
		}

	}

	return ids, nil
}

//GetAllRequests ...
func (r *FriendsRepository) GetAllRequests(userid int) ([]model.Friends, error) {
	var friends []model.Friends

	rows, err := r.store.db.Query("SELECT user1,user2 from friends where (user1 = $1) AND (status = 0 OR status = 2)",
		userid,
	)
	if err != nil {
		return friends, err
	}

	for rows.Next() {
		friend := model.Friends{}
		err := rows.Scan(&friend.User1, &friend.User2)
		if err != nil {
			return friends, err
		}

		friends = append(friends, friend)

	}
	return friends, err
}

//GetAllRequests ...
func (r *FriendsRepository) GetRequests(userid int) ([]model.Friends, error) {
	var friends []model.Friends

	rows, err := r.store.db.Query("SELECT user1,user2 from friends where (user2 = $1) AND (status = 0)",
		userid,
	)
	if err != nil {
		return friends, err
	}

	for rows.Next() {
		friend := model.Friends{}
		err := rows.Scan(&friend.User1, &friend.User2)
		if err != nil {
			return friends, err
		}

		friends = append(friends, friend)

	}
	return friends, err
}

//Get ...
func (r *FriendsRepository) Get(userid int) ([]model.Friends, error) {
	var friends []model.Friends

	rows, err := r.store.db.Query("SELECT user1,user2,status from friends where (user1 = $1 OR user2 = $1) AND status=1",
		userid,
	)
	if err != nil {
		return friends, err
	}

	for rows.Next() {
		friend := model.Friends{}
		err := rows.Scan(&friend.User1, &friend.User2, &friend.Status)
		if err != nil {
			return friends, err
		}

		friends = append(friends, friend)

	}
	return friends, err
}

//GetStatusFriend ...
func (r *FriendsRepository) GetStatusFriend(f *model.Friends) error {

	var userid2 int

	err := r.store.db.QueryRow("SELECT status,user2 from friends where (user1 = $1 and user2= $2) OR (user1 = $2 and user2= $1)",
		f.User1,
		f.User2,
	).Scan(&f.Status, &userid2)

	f.ForMe = f.User1 == userid2 // Если совпадает то заявка была отправлена первому пользователю

	if err != nil && err.Error() == "sql: no rows in result set" {
		f.Status = 3
		return nil
	}

	return err
}

//GetAllFriends ...
func (r *FriendsRepository) GetAllFriends(user int) ([]model.Friends, error) {
	return []model.Friends{}, nil
}

//GetAllSubscribes ...
func (r *FriendsRepository) GetAllSubscribes(user int) ([]model.Friends, error) {
	var friends []model.Friends
	rows, err := r.store.db.Query("SELECT user1,user2 from friends where (user2 = $1) AND (status = 0 OR status = 2)",
		user,
	)
	if err != nil {
		return friends, err
	}

	for rows.Next() {
		friend := model.Friends{}
		err := rows.Scan(&friend.User1, &friend.User2)
		if err != nil {
			return friends, err
		}

		friends = append(friends, friend)

	}

	return friends, nil
}
