package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ernado/selectel/storage"
	"github.com/go-redis/redis"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/katelinlis/UserBackend/internal/app/store"
	"github.com/sirupsen/logrus"
)

type server struct {
	router *mux.Router
	logger *logrus.Logger
	store  store.Store
	redis  *redis.Client
}

const (
	ctxKeyUser ctxKey = iota
)

type ctxKey int8

var (
	errIncorrectEmailOrPassword = errors.New("incorect email or password")
	jwtsignkey                  string
	recaptchaSecret             string
	selectelUser                string
	selectelPassword            string
)

func newServer(store store.Store, config *Config) *server {
	s := &server{
		router: mux.NewRouter(),
		logger: logrus.New(),
		redis: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		store: store,
	}
	s.configureRouter()

	jwtsignkey = config.JwtSignKey
	recaptchaSecret = config.RecaptchaSecret
	selectelUser = config.Selectel.User
	selectelPassword = config.Selectel.Password

	return s
}

func (s *server) UploadSelectel(reader io.Reader, url string) string {
	api, err := storage.New(selectelUser, selectelPassword)
	if err != nil {
		log.Fatal(err)
	}

	conteiner := api.Container("Social")
	fmt.Print(conteiner.Info())
	err = conteiner.Upload(reader, url, "")
	if err != nil {
		log.Fatal(err)
	}
	return ""
}

func (s *server) GetDataFromToken(w http.ResponseWriter, r *http.Request) (float64, error) {
	var token string
	tokens, ok := r.Header["Authorization"]
	if ok && len(tokens) >= 1 {
		token = tokens[0]
		token = strings.TrimPrefix(token, "Bearer ")
	}

	if token == "" {
		return 0, errors.New("Token is missing")
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			msg := fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			return 0, msg
		}
		return []byte(jwtsignkey), nil
	})

	if err != nil {
		//s.error(w, r, http.StatusUnauthorized, errors.New("Error parsing token"))
		return 0, errors.New("Error parsing token")
	}
	if parsedToken != nil && parsedToken.Valid {
		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
			userid := claims["userid"].(float64)
			return userid, nil
		}
	}
	return 0, nil

}

//UrlLimitOffset ...
func (s *server) URLLimitOffset(request *http.Request) (int, int) {
	var offsetVar int
	var limitVar = 20
	urlParams := request.URL.Query()
	if len(urlParams["offset"]) > 0 {
		offset, err := strconv.Atoi(urlParams["offset"][0])
		if err == nil {
			offsetVar = offset
		}
	}
	if len(urlParams["limit"]) > 0 {
		limit, err := strconv.Atoi(urlParams["limit"][0])
		if err == nil {
			limitVar = limit
		}
	}
	return offsetVar, limitVar
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, POST, GET, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")

	defer r.Body.Close()
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {

	s.router.Use(s.loggingMiddleware)

	s.router.Methods("OPTIONS").HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Access-Control-Allow-Origin", "*")
			rw.Header().Set("Access-Control-Allow-Methods", "DELETE, POST, GET, PUT, OPTIONS")
			rw.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
			rw.WriteHeader(http.StatusOK)
		})

	s.ConfigureWallRouter()
	s.ConfigureUserRouter()
}

func (s *server) VerifyRecaptcha(captcha string) bool {

	data := url.Values{
		"secret":   {recaptchaSecret},
		"response": {captcha},
	}

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", data)
	if err != nil {
		log.Fatalln(err)
	}

	type Response struct {
		ChallengeTS string `json:"challenge_ts"`
		Hostname    string `json:"hostname"`
		Success     bool   `json:"success"`
	}

	var result Response
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println(result)

	return result.Success
}

func (s *server) emptyresponse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug("Request " + r.RequestURI + " from " + r.RemoteAddr)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
	return
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)

	r.Body.Close()

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
