package main

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"site1/cmd/database"
	"strings"
)

func init() {

	if os.Getenv("ENV") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

func sendHeader(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Authorization, Version, Mode, RequestUrl")
		w.Header().Add("Content-type", "application/json; charset=utf-8")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
		return
	})
}

func jwtToken(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip Auth for OPTIONS
		if r.Method == "OPTIONS" || r.URL.Path == "/register" || r.URL.Path == "/login" || r.URL.Path == "/" {
			next.ServeHTTP(w, r)
			return
		}
		r.Header.Set("USER_ID", "")

		// authentification and check Authorization Header
		var tokenString = r.Header.Get("Authorization")
		if tokenString != "" {
			if strings.Index(tokenString, "Bearer ") == 0 {
				tokenString = tokenString[len("Bearer "):]
			}
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v.", token.Header["alg"])
				}

				return []byte(os.Getenv("JWT_SECRET")), nil
			})

			var tokenInfo map[string]interface{}

			if err == nil {
				if i, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					tokenInfo = i
				} else {
					unauthorized(w, errors.New("Invalid token"))
				}
			} else {
				unauthorized(w, errors.New("Invalid token"))
			}

			user, err := User{}.FindById(uint(tokenInfo["id"].(float64)))
			//timer["jwtMiddleware_1_aq"] = time.Since(start).String()
			if err != nil || user.ID == 0 {
				unauthorized(w, errors.New("Invalid token"))
				return
			}
			r.Header.Set("USER_ID", fmt.Sprintf("%d", user.ID))
			next.ServeHTTP(w, r)
			return
		}
		unauthorized(w, errors.New("Invalid token"))
		return
	})
}

func main() {
	var router = mux.NewRouter()

	db, _ := database.Connect()
	db = db.Debug()
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Ad{})

	router.Use(sendHeader)
	router.Use(jwtToken)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json; charset=utf-8")
		w.Write([]byte(`{"result":"ok", "site": 5}`))
	}).Methods("GET")

	router.HandleFunc("/test_db", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json; charset=utf-8")
	}).Methods("GET")

	router.HandleFunc("/register", register).Methods("POST")
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/api/users/{id}", getUser).Methods("GET")
	router.HandleFunc("/api/users/{id}/ads", getUserAds).Methods("GET")
	router.HandleFunc("/api/users/{id}/password", updatePassword).Methods("PUT")

	router.HandleFunc("/api/ads", createAd).Methods("POST")
	router.HandleFunc("/api/ads", getAllAd).Methods("GET")
	router.HandleFunc("/api/ads/{id}", getAd).Methods("GET")
	router.HandleFunc("/api/ads/{id}", updateAd).Methods("PUT")

	e := http.ListenAndServe(":"+os.Getenv("INTERNAL_PORT"), router)
	if e != nil {
		log.Fatal("ListenAndServe: ", e)
	}
}
