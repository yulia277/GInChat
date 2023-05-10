package main


import (
	"IMProject/controller"
	"IMProject/utils"
	"github.com/spf13/viper"
)
func main() {
	utils.InitConfig()
	utils.InitMySQL()
	utils.InitRedis()

	r := controller.Router()
	r.Run(viper.GetString("port.server"))
}