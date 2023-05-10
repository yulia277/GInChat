package controller

import (
	"IMProject/docs"
	"IMProject/service"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()

	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	r.Static("/asset", "asset/")
	r.LoadHTMLGlob("views/**/*")

	//首页
	r.GET("/index", service.GetIndex)
	r.GET("/toRegister",service.ToRegister)
	r.GET("/toChat", service.ToChat)
	r.GET("/chat", service.Chat)
	r.POST("/searchFriends", service.SearchFriends)
	//用户模块
	r.POST("/user/getUserList", service.GetUserList)
	r.POST("/user/createUser", service.CreateUser)
	r.POST("/user/deleteUser", service.DeleteUser)
	r.POST("/user/updateUser", service.UpdateUser)
	r.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)
	r.POST("/user/find", service.FindByID)
	//发送消息
	r.GET("/user/sendMsg", service.SendMsg)
	r.GET("/user/sendUserMsg", service.SendUserMsg)
	//上传文件
	r.POST("/attach/upload", service.Upload)
	//添加好友
	r.POST("/contact/addfriend", service.AddFriends)
	//创建群
	r.POST("/contact/createCommunity", service.CreateCommunity)
	//加载群列表
	r.POST("/contact/loadcommunity",service.LoadCommunity)
	r.POST("/contact/joinGroup", service.JoinGroups)
	r.POST("/user/redisMsg", service.RedisMsg)
	return r
}
