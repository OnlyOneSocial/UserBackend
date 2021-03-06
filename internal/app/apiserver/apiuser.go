package apiserver

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/katelinlis/UserBackend/internal/app/model"
	"github.com/nfnt/resize"
)

func (s *server) ConfigureUserRouter() {

	router := s.router.PathPrefix("/api/user").Subrouter()
	router.HandleFunc("/get/{id}", s.HandleGetUser()).Methods("GET")                  // Получение данных о пользователе
	router.HandleFunc("/get", s.HandleGetUsers()).Methods("GET")                      // Получение данных о пользователях
	router.HandleFunc("/login", s.HandleLoginUser()).Methods("POST")                  // Авторизация
	router.HandleFunc("/register", s.HandleCreateUser()).Methods("POST")              // Регистрация
	router.HandleFunc("/settings", s.HandleChangeSettingsMain()).Methods("PUT")       /* Изменение настроек*/
	router.HandleFunc("/status", s.HandleChangeStatus()).Methods("PUT")               /* Изменение статуса*/
	router.HandleFunc("/password", s.HandleChangePassword()).Methods("PUT")           // Изменение пароля
	router.HandleFunc("/banlist/{id}", s.HandleAddToBanlist()).Methods("POST")        // Заблокировать пользователя
	router.HandleFunc("/banlist/{id}", s.HandleRemoveFromBanlist()).Methods("DELETE") // Заблокировать пользователя
	router.HandleFunc("/avatar", s.HandleChangeAvatar()).Methods("PUT")               // Сменить аватар
}

func (s *server) JWTproccessingAndUpdateOnline(w http.ResponseWriter, request *http.Request) (int, error) {
	var id int

	userid, err := s.GetDataFromToken(w, request)

	if err != nil {
		return 0, err
	}

	id = int(userid)

	s.store.User().UpdateOnline(id)

	return id, nil

}

func (s *server) HandleChangeAvatar() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		userid, err := s.JWTproccessingAndUpdateOnline(w, request)
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
		}
		if userid <= 0 {
			s.error(w, request, http.StatusBadRequest, err)
		}

		request.ParseMultipartForm(10 << 20)

		file, handler, err := request.FormFile("myFile")
		if err != nil {
			fmt.Println("Error Retrieving the File")
			fmt.Println(err)
			return
		}
		//defer file.Close()
		fmt.Printf("Uploaded File: %+v\n", handler.Filename)
		fmt.Printf("File Size: %+v\n", handler.Size)
		fmt.Printf("MIME Header: %+v\n", handler.Header)
		fmt.Println(handler.Header.Get("Content-Type"))

		orgImage, _, err := image.Decode(file)
		// check err

		newImage := resize.Thumbnail(100, 100, orgImage, resize.Lanczos3)
		var StringTypeFile string
		buf := new(bytes.Buffer)
		if handler.Header.Get("Content-Type") == "image/jpeg" {
			err = jpeg.Encode(buf, newImage, nil)
			StringTypeFile = "jpg"
		}
		if handler.Header.Get("Content-Type") == "image/png" {
			err = png.Encode(buf, newImage)
			StringTypeFile = "png"
		}
		sendS3 := buf.Bytes()

		buf = new(bytes.Buffer)
		if handler.Header.Get("Content-Type") == "image/jpeg" {
			err = jpeg.Encode(buf, orgImage, nil)
		}
		if handler.Header.Get("Content-Type") == "image/png" {
			err = png.Encode(buf, orgImage)
		}
		sendS3Orig := buf.Bytes()
		hash := md5.Sum(sendS3Orig)
		var md5Name = hex.EncodeToString(hash[:])

		if handler.Header.Get("Content-Type") == "image/jpeg" || handler.Header.Get("Content-Type") == "image/png" {

			s.UploadSelectel(bytes.NewReader(sendS3Orig), "/public/clients/"+strconv.Itoa(userid)+"/"+md5Name+"."+StringTypeFile)
			s.UploadSelectel(bytes.NewReader(sendS3), "/public/clients/"+strconv.Itoa(userid)+"/100-"+md5Name+"."+StringTypeFile)
			s.store.User().SetAvatar(userid, md5Name+"."+StringTypeFile)
		}

		fmt.Fprintf(w, "Successfully Uploaded File\n")
	}
}

func (s *server) HandleRemoveFromBanlist() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {

		userid, err := s.JWTproccessingAndUpdateOnline(w, request)
		if err != nil {
			fmt.Println(err)
		}

		vars := mux.Vars(request)
		whoToBanned, err := strconv.Atoi(vars["id"])
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
			return
		}
		_, err = s.store.User().Find(whoToBanned)
		if err != nil {
			s.error(w, request, http.StatusNotFound, errors.New("not found"))
			return
		}

		status, err := s.store.User().RemoveFromBanList(userid, whoToBanned)
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
			return
		}
		if !status {
			s.respond(w, request, http.StatusUnprocessableEntity, status)
			return
		}
		s.respond(w, request, http.StatusOK, "ok")

	}
}

func (s *server) HandleAddToBanlist() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {

		userid, err := s.JWTproccessingAndUpdateOnline(w, request)
		if err != nil {
			fmt.Println(err)
		}

		vars := mux.Vars(request)
		whoToBanned, err := strconv.Atoi(vars["id"])
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
			return
		}
		_, err = s.store.User().Find(whoToBanned)
		if err != nil {
			s.error(w, request, http.StatusNotFound, errors.New("not found"))
			return
		}

		status, err := s.store.User().AddToBanList(userid, whoToBanned)
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
			return
		}
		if !status {
			s.respond(w, request, http.StatusUnprocessableEntity, status)
			return
		}
		s.respond(w, request, http.StatusOK, "ok")

	}
}

func (s *server) HandleChangeSettingsMain() http.HandlerFunc {
	type ChangeSettings struct {
		BirthDayDate int    `json:"birthday"`
		Gender       string `json:"gender"`
		Country      string `json:"country"`
		City         string `json:"city"`
		Bio          string `json:"bio"`
	}
	return func(w http.ResponseWriter, request *http.Request) {
		userid, err := s.JWTproccessingAndUpdateOnline(w, request)
		if err != nil {
			fmt.Println(err)
		}

		var changeSettings ChangeSettings
		json.NewDecoder(request.Body).Decode(&changeSettings)

		settings := model.SettingsMain{
			ID:           userid,
			BirthdayDate: changeSettings.BirthDayDate,
			Gender:       changeSettings.Gender,
			Country:      changeSettings.Country,
			City:         changeSettings.City,
			Bio:          changeSettings.Bio,
		}

		s.store.User().ChangeSettingsMain(&settings)

		s.respond(w, request, http.StatusOK, settings)
	}
}

func (s *server) HandleChangePassword() http.HandlerFunc {
	type ChangePassword struct {
		OldPassword string `json:"OldPassword"`
		Password    string `json:"password"`
	}
	return func(w http.ResponseWriter, request *http.Request) {
		userid, err := s.JWTproccessingAndUpdateOnline(w, request)
		if err != nil {
			fmt.Println(err)
		}

		var changePassword ChangePassword
		json.NewDecoder(request.Body).Decode(&changePassword)

		user, err := s.store.User().Find(userid)

		user, err = s.store.User().FindByUsername(user.Username)

		if !user.ComparePassword(changePassword.OldPassword) {
			s.error(w, request, http.StatusBadRequest, errors.New("password incorrect"))
			return
		}

		user.Password = changePassword.Password
		s.store.User().ChangePassword(&user)

		s.respond(w, request, http.StatusOK, user)
	}
}

func (s *server) HandleChangeStatus() http.HandlerFunc {
	type ChangeStatus struct {
		Status string `json:"status"`
	}
	return func(w http.ResponseWriter, request *http.Request) {
		userid, err := s.JWTproccessingAndUpdateOnline(w, request)
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
			return
		}

		var changeStatus ChangeStatus
		json.NewDecoder(request.Body).Decode(&changeStatus)

		s.store.User().ChangeStatus(userid, changeStatus.Status)

		s.respond(w, request, http.StatusOK, userid)
	}
}

func (s *server) HandleGetUser() http.HandlerFunc {
	type FriendStatus struct {
		ForMe  bool `json:"forme"`
		Status int  `json:"status"`
	}
	type User struct {
		FriendStatus FriendStatus  `json:"friend_status"`
		User         model.User    `json:"user"`
		Friends      FriendsStruct `json:"friends"`
	}
	return func(w http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		userid2, err := strconv.Atoi(vars["id"])
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
			return
		}

		user, err := s.store.User().Find(userid2)
		if err != nil {
			s.error(w, request, http.StatusNotFound, errors.New("not found"))
			return
		}

		userid, err := s.JWTproccessingAndUpdateOnline(w, request)
		if int(userid) > 0 {
			user.Me = userid2 == userid
		}

		friends, err := s.GetFriends(userid2)
		if err != nil {
			return
		}

		friendStatus := model.Friends{}
		friendStatus.User1 = int(userid)
		friendStatus.User2 = userid2

		err = s.store.Friends().GetStatusFriend(&friendStatus)
		if err != nil && err.Error() == "sql: no rows in result set" {
			friendStatus.Status = 3
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		UserModel := User{
			User:    user,
			Friends: friends,
			FriendStatus: FriendStatus{
				ForMe:  friendStatus.ForMe,
				Status: friendStatus.Status,
			},
		}

		s.respond(w, request, http.StatusOK, UserModel)
	}
}

func (s *server) HandleGetUsers() http.HandlerFunc {
	type Users struct {
		Total int          `json:"total"`
		Users []model.User `json:"users"`
	}
	return func(w http.ResponseWriter, request *http.Request) {
		users, err := s.store.User().Get()
		if err != nil {
			return
		}

		count, err := s.store.User().GetCount()
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		s.respond(w, request, http.StatusOK, Users{Users: users, Total: count})
	}
}

//InitJWT ...
func (s *server) InitJWT(UserID int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := make(jwt.MapClaims)
	claims["userid"] = UserID
	claims["exp"] = time.Now().Add(time.Hour * 6).Unix()
	token.Claims = claims
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(jwtsignkey))
	return tokenString, err
}

//Register ...
type Register struct {
	Password  string `json:"password"`
	Login     string `json:"username"`
	Recaptcha string `json:"captcha"`
}

func (s *server) HandleCreateUser() http.HandlerFunc {
	type UserLoginOutput struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
		Jwt      string `json:"jwt"`
	}
	return func(w http.ResponseWriter, request *http.Request) {
		var createPost Register
		json.NewDecoder(request.Body).Decode(&createPost)

		var user model.User
		user.Password = createPost.Password
		user.Username = createPost.Login

		status := s.VerifyRecaptcha(createPost.Recaptcha)

		if status != true {
			s.error(w, request, http.StatusBadRequest, errors.New("captcha is incorrect"))
			return
		}

		username := strings.ToLower(createPost.Login)
		username = strings.TrimSpace(username)

		if username == "" {
			s.error(w, request, http.StatusBadRequest, errors.New("username is empty"))
			return
		}

		userFind, err := s.store.User().FindByUsername(username)

		if err != nil && err.Error() != "sql: no rows in result set" {
			s.error(w, request, http.StatusBadRequest, err)
			return
		}

		if userFind.Username != "" {
			s.error(w, request, http.StatusBadRequest, errors.New("this user is registered"))
			return
		}

		err = s.store.User().Create(&user)
		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
			return
		}

		jwt, err := s.InitJWT(user.ID)

		if err != nil {
			s.error(w, request, http.StatusBadRequest, err)
			return
		}

		Output := UserLoginOutput{ID: user.ID, Username: user.Username, Avatar: user.Avatar, Jwt: jwt}
		s.respond(w, request, http.StatusOK, Output)
	}
}

func (s *server) HandleLoginUser() http.HandlerFunc {
	type UserLoginOutput struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
		Jwt      string `json:"jwt"`
	}
	return func(w http.ResponseWriter, request *http.Request) {
		var createPost Register
		json.NewDecoder(request.Body).Decode(&createPost)

		var user model.User
		user.Password = createPost.Password

		status := s.VerifyRecaptcha(createPost.Recaptcha)

		if status != true {
			s.error(w, request, http.StatusBadRequest, errors.New("captcha is incorrect"))
			return
		}

		username := strings.ToLower(createPost.Login)
		username = strings.TrimSpace(username)
		user.Username = username

		if username == "" {
			s.error(w, request, http.StatusBadRequest, errors.New("username is empty"))
			return
		}

		userFind, err := s.store.User().FindByUsername(username)

		if err != nil && err.Error() == "sql: no rows in result set" {
			s.error(w, request, http.StatusBadRequest, errors.New("this user is not found"))
			return
		}

		if !userFind.ComparePassword(createPost.Password) {
			s.error(w, request, http.StatusBadRequest, errors.New("password is incorrect"))
			return
		}

		if userFind.ComparePassword(createPost.Password) {

			jwt, err := s.InitJWT(userFind.ID)

			if err != nil {
				s.error(w, request, http.StatusBadRequest, err)
				return
			}

			Output := UserLoginOutput{ID: userFind.ID, Username: userFind.Username, Avatar: userFind.Avatar, Jwt: jwt}
			s.respond(w, request, http.StatusOK, Output)
		}

	}
}
