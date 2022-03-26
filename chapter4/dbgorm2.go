package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	//	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"helloworld/chapter4/model"
	"log"
	"os"
)

var (
	db            *gorm.DB
	sqlConnection = "root:Jzyz.8888@(192.168.8.151:3306)/chapter4?" + "charset=utf8mb4&parseTime=true"
)

func init() {
	var err error
	db, err = gorm.Open("mysql", sqlConnection)
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&model.GormUser{})
}

func main() {
	defer db.Close()
	/*
		//创建用户
		GormUser:=model.GormUser{
			Phone: "13600000000",
			Name: "Shirdon",
			Password: md5Password("111111"),  //用户密码
		}
		db.Save(&GormUser)    //保存到数据库
		//db.Create(&GormUser)   //保存到数据库


		//查询用户

		var GormUser=new(model.GormUser)
		db.Where("phone = ?","13600000000").Find(&GormUser)
		db.First(&GormUser, "phone = ?","13600000000")
		fmt.Println(GormUser)


		//更新用户
		var GormUser=new(model.GormUser)
		err:=db.Model(&GormUser).Where("phone = ?","13600000000").Update("phone","13800000000").Error
		if err!=nil{
			fmt.Println(err)
		}

		//删除用户
		var GormUser=new(model.GormUser)
		db.Where("phone = ?","13600000000").Delete(&GormUser)
	*/

	//开启事务
	tx := db.Begin()
	GormUser := model.GormUser{
		Phone:    "13600000000",
		Name:     "Shirdon",
		Password: md5Password("111111"), //用户密码
	}
	if err := tx.Create(&GormUser).Error; err != nil {
		tx.Rollback()
		fmt.Println()
	}
	db.First(&GormUser, "phone = ?", "13600000000")
	tx.Commit()
	/*
	 */

	db.LogMode(true)
	db.SetLogger(log.New(os.Stdout, "\r\n", 0))

}

func md5Password(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
