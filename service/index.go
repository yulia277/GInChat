package service

import (
	"IMProject/dao"
	"github.com/gin-gonic/gin"
)



//GetIndex
//@Tags 首页
//@Success 200 {string} welcome
//@Router /index [get]
func GetIndex(c *gin.Context) {
	//ind, err := template.ParseFiles("index.html", "views/chat/head.html")
	//if err != nil {
	//	fmt.Println("渲染不成功")
	//}
	//ind.Execute(c.Writer, "index")
	//if err != nil {
	//	fmt.Println("渲染不成功")
	//	panic(err)
	//}
	c.HTML(200 , "index.html", nil)
}

func ToRegister(c *gin.Context) {
	c.HTML(200 , "register.html", nil)
}


func ToChat(c *gin.Context) {
	c.HTML(200 , "/chat/index.shtml", nil)
}

func Chat(c *gin.Context) {
	dao.Chat(c.Writer, c.Request)
}