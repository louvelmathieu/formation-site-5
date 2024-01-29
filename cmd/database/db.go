package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

func Connect() (*gorm.DB, error) {
	var c = mysql.Open(os.Getenv("RDS_USERNAME") + ":" + os.Getenv("RDS_PASSWORD") + "@tcp(" + os.Getenv("RDS_HOSTNAME") + ":" + os.Getenv("RDS_PORT") + ")/" + os.Getenv("RDS_DB_NAME") + "?charset=utf8&parseTime=True&loc=UTC")

	db, err := gorm.Open(c, &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	db = db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8 auto_increment=1")
	return db.Session(&gorm.Session{}), nil
}
