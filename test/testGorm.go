package main

import (
	"IMProject/dao"
	"IMProject/utils"
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	utils.InitConfig()
	//连接数据库：用户名:密码@协议(IP:port)/数据库名?
	//db, err := gorm.Open("mysql", "root:888888@(127.0.0.1:3306)/ginchat?charset=utf8mb4&parseTime=True&loc=Local")
	fmt.Println(viper.GetString("======mysql.dns====="))
	db, err :=gorm.Open(mysql.Open(viper.GetString("mysql.dns")),&gorm.Config{})

	if err != nil {
		panic(err)
	}
	//自动迁移
	//db.AutoMigrate(&dao.Message{})
	db.AutoMigrate(&dao.GroupBasic{})
	//db.AutoMigrate(&dao.Contact{})

	//插入
	//user := &dao.UserBasic{}
	//user.Name = "dazai1"
	//user.Password = "5678"
	//user.LoginTime = time.Now()
	//user.HeartbeatTime = time.Now()
	//user.LoginOutTime = time.Now()
	//db.Create(user)
}
