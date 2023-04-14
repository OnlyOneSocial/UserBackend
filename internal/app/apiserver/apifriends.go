package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/katelinlis/UserBackend/internal/app/model"
)

func (s *server) ConfigureWallRouter() {

	router := s.router.PathPrefix("/api/friends").Subrouter()
	router.HandleFunc("/request/{id}", s.HandleRequest()).Methods("POST")              // Отправить заявку
	router.HandleFunc("/request_accept/{id}", s.HandleRequestAccept()).Methods("POST") // Принять заявку
	router.HandleFunc("/request_cancel", s.HandleHerobrine()).Methods("GET")           // Отменить заявку
	router.HandleFunc("/delete/{id}", s.HandleHerobrine()).Methods("DELETE")           // Удаление человека из друзей
	router.HandleFunc("/array_friends/{id}", s.HandleGetArray()).Methods("GET")        // получение array друзей для бэкенда стены
	router.HandleFunc("/get/{id}", s.HandleGet()).Methods("GET")                       // получение списка друзей пользователя
	router.HandleFunc("/request", s.HandleGetRequests()).Methods("GET")                // получение списка заявок пользователю
}

func (s *server) HandleGet() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		userid, err := strconv.Atoi(vars["id"])
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
		}

		friends, err := s.GetFriends(userid)
		if err != nil {
			return
		}

		s.respond(w, request, http.StatusOK, friends)
	}
}

//FriendsStruct ...
type FriendsStruct struct {
	Count int             `json:"count"`
	List  []model.Friends `json:"list"`
}

func (s *server) GetFriends(userid int) (FriendsStruct, error) {

	friends, err := s.store.Friends().Get(userid)
	if err != nil {
		return FriendsStruct{
			List: friends, Count: 0,
		}, err
	}

	for index, friend := range friends {
		var friendID int
		if friend.User1 == userid {
			friendID = friend.User2
		}

		if friend.User2 == userid {
			friendID = friend.User1
		}

		user, err := s.store.User().Find(friendID)

		if err != nil {
			return FriendsStruct{
				List: friends, Count: 0,
			}, err
		}

		friends[index].User = user
	}

	count, err := s.store.Friends().GetCount(userid)
	if err != nil {
		return FriendsStruct{
			List: friends, Count: count,
		}, err
	}

	return FriendsStruct{
		List: friends, Count: count,
	}, err

}

func (s *server) GetRequests(userid int) (FriendsStruct, error) {

	friends, err := s.store.Friends().GetRequests(userid)
	if err != nil {
		return FriendsStruct{
			List: friends, Count: 0,
		}, err
	}

	for index, friend := range friends {
		var friendID int
		if friend.User1 == userid {
			friendID = friend.User2
		}

		if friend.User2 == userid {
			friendID = friend.User1
		}

		user, err := s.store.User().Find(friendID)

		if err != nil {
			return FriendsStruct{
				List: friends, Count: 0,
			}, err
		}

		friends[index].User = user
	}

	count, err := s.store.Friends().GetCountRequests(userid)
	if err != nil {
		return FriendsStruct{
			List: friends, Count: count,
		}, err
	}

	return FriendsStruct{
		List: friends, Count: count,
	}, err

}

func (s *server) HandleGetRequests() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		userid, err := s.GetDataFromToken(w, request)

		if err != nil {
			s.error(w, request, http.StatusUnauthorized, err)
			return
		}

		Requests, err := s.GetRequests(int(userid.LegacyID))

		if err != nil {
			s.error(w, request, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, request, http.StatusOK, Requests)
	}
}

func (s *server) HandleGetSubscribes() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		userid, err := s.GetDataFromToken(w, request)

		if err != nil {
			s.error(w, request, http.StatusUnauthorized, err)
			return
		}

		Requests, err := s.store.Friends().GetAllRequests(int(userid.LegacyID))

		if err != nil {
			s.error(w, request, http.StatusUnprocessableEntity, err)
			return
		}

		s.respond(w, request, http.StatusOK, Requests)
	}
}

func (s *server) HandleGetArray() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		userid, err := strconv.Atoi(vars["id"])
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
		}

		friends, err := s.store.Friends().GetArrayFriends(userid)
		if err != nil {
			return
		}

		s.respond(w, request, http.StatusOK, friends)
	}
}

func (s *server) HandleRequest() http.HandlerFunc {
	type Request struct {
		UserID int `json:"id"`
	}
	return func(w http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		userid2, err := strconv.Atoi(vars["id"])
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
		}

		userid, err := s.GetDataFromToken(w, request)

		if err != nil {
			s.error(w, request, http.StatusUnauthorized, err)
			return
		}

		friend := model.Friends{
			User1: int(userid.LegacyID),
			User2: userid2,
		}

		err = s.store.Friends().SendRequest(&friend)

		if err != nil {
			s.error(w, request, http.StatusUnprocessableEntity, err)
			return
		}

		b := &bytes.Buffer{}

		thisUser, err := s.store.User().Find(int(userid.LegacyID))

		reqData := model.Notification{
			Type:      "request_send",
			UserID:    thisUser.ID,
			UserName:  thisUser.Username,
			TimeStamp: time.Now().Unix(),
		}
		json.NewEncoder(b).Encode(&reqData)

		RequestNotifServer, err := http.Post("http://localhost:3078/api/notification/user/"+strconv.Itoa(userid2), "application/json", b)
		if err != nil {
			s.error(w, request, http.StatusUnprocessableEntity, err)
			fmt.Println(err)
			return
		}

		body, err := ioutil.ReadAll(RequestNotifServer.Body)
		if err != nil {
			fmt.Println(err)
		}

		if string(body) == "ok" {
			fmt.Println("ok")
		}

		s.respond(w, request, http.StatusOK, friend)
	}
}

func (s *server) HandleRequestAccept() http.HandlerFunc {
	type Request struct {
		UserID int `json:"id"`
	}
	return func(w http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		userid2, err := strconv.Atoi(vars["id"])
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
		}

		userid, err := s.GetDataFromToken(w, request)

		if err != nil {
			s.error(w, request, http.StatusUnauthorized, err)
		}

		friend := model.Friends{
			User1: int(userid.LegacyID),
			User2: userid2,
		}

		err = s.store.Friends().Accept(&friend)

		if err != nil {
			s.error(w, request, http.StatusUnprocessableEntity, err)
		}

		b := &bytes.Buffer{}

		thisUser, err := s.store.User().Find(int(userid.LegacyID))

		reqData := model.Notification{
			Type:      "request_accept",
			UserID:    thisUser.ID,
			UserName:  thisUser.Username,
			TimeStamp: time.Now().Unix(),
		}
		json.NewEncoder(b).Encode(&reqData)

		RequestNotifServer, err := http.Post("http://localhost:3078/api/notification/user/"+strconv.Itoa(userid2), "application/json", b)
		if err != nil {
			s.error(w, request, http.StatusUnprocessableEntity, err)
			fmt.Println(err)
			return
		}

		body, err := ioutil.ReadAll(RequestNotifServer.Body)
		if err != nil {
			fmt.Println(err)
		}

		if string(body) == "ok" {
			fmt.Println("ok")
		}

		s.respond(w, request, http.StatusOK, friend)
	}
}

func (s *server) HandleHerobrine() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {

		s.respond(w, request, http.StatusOK, "")
	}
}

//MessageSend ...
type MessageSend struct {
	Text   string `json:"text"`
	SendTo int    `json:"to"`
}
