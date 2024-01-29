package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"site1/cmd/database"
	"strconv"
	"time"
)

type User struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Account information
	Email     string `json:"email" gorm:"type:varchar(255);uniqueIndex"`
	Lastname  string `json:"lastname" gorm:"type:varchar(255)"`
	Firstname string `json:"firstname" gorm:"type:varchar(255)"`
	Phone     string `json:"phone" gorm:"type:varchar(255)"`

	Password     string `json:"password" gorm:"type:text"`
	TextPassword string `json:"text_password" gorm:"-"`
}

func (user User) FindOne(email string) (User, error) {
	db, err := database.Connect()
	if err != nil {
		return user, err
	}

	err = db.Last(&user, map[string]interface{}{
		"email": email,
	}).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, nil
	}

	return user, err
}

func (user User) FindById(id uint) (User, error) {
	db, err := database.Connect()
	if err != nil {
		return user, err
	}

	err = db.Last(&user, map[string]interface{}{
		"id": id,
	}).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return User{}, nil
	}

	return user, err
}

func (user *User) Register() error {
	if user.TextPassword != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(user.TextPassword), bcrypt.MinCost)
		user.Password = string(hash)
	} else {
		return errors.New("Empty password")
	}

	if user.Email == "" {
		return errors.New("Empty email")
	}

	db, err := database.Connect()
	if err != nil {
		return err
	}

	var exist User

	tx := db.Session(&gorm.Session{})
	err = tx.First(&exist, map[string]interface{}{
		"email": user.Email,
	}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Insert
			tx := db.Session(&gorm.Session{})
			err = tx.Create(&user).Error
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		return errors.New("Dupplicate E-mail")
	}

	return nil
}

func (user *User) CheckPassword(plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainPwd))
	if err != nil {
		return false
	}

	return true
}

func (user *User) UpdatePassword(plainPwd string) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte(plainPwd), bcrypt.MinCost)

	db, err := database.Connect()
	if err != nil {
		return err
	}

	return db.Model(&user).Updates(User{
		Password: string(hash),
	}).Error
}

func (user *User) UpdateAccount() error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	var exist User
	err = db.First(&exist, map[string]interface{}{
		"email": user.Email,
	}).Error
	if err != nil {
		return errors.New("Dupplicate E-mail")
	}

	err = db.Model(&user).UpdateColumns(map[string]interface{}{
		"Email":     user.Email,
		"Lastname":  user.Lastname,
		"Firstname": user.Firstname,
		"Phone":     user.Phone,
	}).Error
	if err != nil {
		return err
	}

	return nil
}

func createJWTToken(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":        user.ID,
		"email":     user.Email,
		"firstname": user.Firstname,
		"lastname":  user.Lastname,
		"time":      time.Now().String(),
	})

	str, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return str, err
}

/**
 * REST API
 */
func updatePassword(w http.ResponseWriter, r *http.Request) {
	var password = struct {
		NewPassword string `json:"new_password"`
	}{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&password)
	if err != nil {
		badRequest(w, err)
		return
	}

	params := mux.Vars(r)
	userId, _ := strconv.Atoi(params["id"])

	authUserId, _ := strconv.Atoi(r.Header.Get("USER_ID"))
	if userId != authUserId {
		unauthorized(w, errors.New("Invalid user"))
		return
	}

	user, _ := User{}.FindById(uint(userId))
	user.UpdatePassword(password.NewPassword)

	if err != nil {
		badRequest(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func register(w http.ResponseWriter, r *http.Request) {
	var user User
	bodyBytes, _ := io.ReadAll(r.Body)
	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	err := decoder.Decode(&user)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = user.Register()
	if err != nil {
		badRequest(w, err)
		return
	}

	j, err := json.Marshal(user)
	if err != nil {
		badRequest(w, err)
		return
	}
	w.Write(j)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("USER_ID") == "" {
		return
	}

	params := mux.Vars(r)
	userId, _ := strconv.Atoi(params["id"])

	user, err := User{}.FindById(uint(userId))
	if err != nil || user.ID == 0 {
		unauthorized(w, errors.New("Invalid user"))
		return
	}

	json.NewEncoder(w).Encode(user)
}

func login(w http.ResponseWriter, r *http.Request) {
	auth := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&auth)
	if err != nil {
		badRequest(w, err)
		return
	}

	if auth.Username == "" || auth.Password == "" {
		badRequest(w, errors.New("Missings fields"))
		return
	}

	user, err := User{}.FindOne(auth.Username)
	if err != nil || user.ID == 0 {
		unauthorized(w, errors.New("Invalid credential"))
		return
	}

	if !user.CheckPassword(auth.Password) {
		unauthorized(w, errors.New("Invalid credential"))
		return
	}

	// Calc plan for trial
	token, err := createJWTToken(user)
	if err != nil {
		unauthorized(w, errors.New("Invalid credential"))
		return
	}

	json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
	}{
		Token: token,
	})
}
