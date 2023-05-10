package utils

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)
var (
	DB *gorm.DB
	Red *redis.Client
	)
func InitConfig() {
	//Viper主要是用于处理各种格式的配置文件，简化程序配置的读取问题。
	viper.SetConfigName("app") //配置文件的名字
	viper.AddConfigPath("config") //配置文件的路径
	err := viper.ReadInConfig() //查找并读取配置文件
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("config app inited")
}

//打开数据库
func InitMySQL() {
	//打印MYSQL日志
	newLogger :=logger.New(
		//自定义日志模板，打印sql语句
		log.New(os.Stdout,"\r\n",log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, //慢sql阈值
			LogLevel: logger.Info,//级别
			Colorful: true,//彩色
		},
		)

	DB, _ =gorm.Open(mysql.Open(viper.GetString("mysql.dns")),&gorm.Config{Logger: newLogger})
	fmt.Println("MySQL inited")
}

func InitRedis() {
	Red = redis.NewClient(&redis.Options{
		Addr : viper.GetString("redis.addr"),
	})
}

const (
	PublishKey = "websocket"

)

//发布消息到redis中，将信息发送给指定的channel
func Publish(ctx context.Context, channel string , msg string) error {
	fmt.Println("Publish ...", msg)
	var err error
	err = Red.Publish(ctx, channel, msg).Err()
	if err != nil {
		fmt.Println(err)
	}
	return err
}


//订阅消息，订阅channel
func Subscribe(ctx context.Context, channel string) (string, error) {

	fmt.Println("Subscribe ...")
	sub := Red.Subscribe(ctx, channel)
	fmt.Println("Subscribe ...",ctx)
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		fmt.Println(err)
		return "",err
	}
	fmt.Println("Subscribe...", msg.Payload)
	return msg.Payload, err
}



















