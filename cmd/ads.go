package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"io"
	"net/http"
	"site1/cmd/database"
	"strconv"
)

type Ad struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	User   User `json:"-" gorm:"ForeignKey:UserId"`
	UserId uint `json:"user_id"`

	Title       string `json:"title" gorm:"type:text"`
	Description string `json:"description" gorm:"type:text"`
	Price       uint   `json:"price" gorm:"type:int"`
}

func (ad *Ad) Create() error {

	db, err := database.Connect()
	if err != nil {
		return err
	}

	return db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&ad).Error
}

func (ad Ad) GetAd(id uint) (Ad, error) {
	db, err := database.Connect()
	if err != nil {
		return ad, err
	}

	err = db.First(&ad, map[string]interface{}{
		"id": id,
	}).Error

	return ad, err
}

func (ad Ad) FindById(userId uint) ([]Ad, error) {
	ads := []Ad{}

	db, err := database.Connect()
	if err != nil {
		return ads, err
	}

	db = db.Omit("")
	err = db.
		Order("id DESC").Find(&ads, map[string]interface{}{
		"user_id": userId,
	}).Error
	if err != nil {
		return []Ad{}, err
	}

	return ads, nil
}

func (ad Ad) FindAll() ([]Ad, error) {
	ads := []Ad{}

	db, err := database.Connect()
	if err != nil {
		return ads, err
	}

	db = db.Omit("")
	err = db.
		Order("id DESC").Find(&ads).Error
	if err != nil {
		return []Ad{}, err
	}

	return ads, nil
}

func (ad *Ad) Update() error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	var exist Ad
	err = db.First(&exist, map[string]interface{}{
		"id": ad.ID,
	}).Error

	if err != nil {
		return err
	}

	err = db.Debug().Model(&exist).Updates(ad).Error
	if err != nil {
		return err
	}

	return nil
}

func getUserAds(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	adId, _ := strconv.Atoi(params["id"])

	ad, err := Ad{}.FindById(uint(adId))
	if err != nil {
		notFound(w, err)
		return
	}

	j, err := json.Marshal(ad)
	if err != nil {
		badRequest(w, err)
		return
	}
	w.Write(j)
}

func getAd(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	adId, _ := strconv.Atoi(params["id"])

	ad, err := Ad{}.GetAd(uint(adId))
	if err != nil {
		notFound(w, err)
		return
	}

	user, _ := User{}.FindById(ad.UserId)

	adFull := struct {
		ID   uint `json:"id"`
		User User `json:"user"`

		Title       string `json:"title" gorm:"type:text"`
		Description string `json:"description" gorm:"type:text"`
		Price       uint   `json:"price" gorm:"type:int"`
	}{
		ID:          ad.ID,
		User:        user,
		Title:       ad.Title,
		Description: ad.Description,
		Price:       ad.Price,
	}

	j, err := json.Marshal(adFull)
	if err != nil {
		badRequest(w, err)
		return
	}

	w.Write(j)
}

func updateAd(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	adId, _ := strconv.Atoi(params["id"])

	var ad Ad
	bodyBytes, _ := io.ReadAll(r.Body)
	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	err := decoder.Decode(&ad)
	if err != nil {
		badRequest(w, err)
		return
	}

	adD, err := Ad{}.GetAd(uint(adId))
	if err != nil {
		notFound(w, err)
		return
	}

	authUserId, _ := strconv.Atoi(r.Header.Get("USER_ID"))
	if adD.UserId != uint(authUserId) {
		unauthorized(w, errors.New("access denied"))
		return
	}

	ad.ID = uint(adId)
	ad.UserId = uint(authUserId)
	ad.Update()

	j, err := json.Marshal(ad)
	if err != nil {
		badRequest(w, err)
		return
	}

	w.Write(j)
}

func getAllAd(w http.ResponseWriter, r *http.Request) {
	ads, _ := Ad{}.FindAll()

	j, err := json.Marshal(ads)
	if err != nil {
		badRequest(w, err)
		return
	}
	w.Write(j)
}

func createAd(w http.ResponseWriter, r *http.Request) {
	var ad Ad
	bodyBytes, _ := io.ReadAll(r.Body)
	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	err := decoder.Decode(&ad)
	if err != nil {
		badRequest(w, err)
		return
	}

	if ad.Title == "" || ad.Price == 0 {
		badRequest(w, errors.New("Empty price or title"))
		return
	}

	authUserId, _ := strconv.Atoi(r.Header.Get("USER_ID"))
	ad.UserId = uint(authUserId)
	ad.Create()

	j, err := json.Marshal(ad)
	if err != nil {
		badRequest(w, err)
		return
	}
	w.Write(j)
}
