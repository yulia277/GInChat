package service

import (
	"IMProject/dao"
	"IMProject/utils"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)


//@Summary 所有用户
//@Tags 用户模块
//@Success 200 {string} json {"code", "message"}
//@Router /user/getUserList [post]
func GetUserList(c *gin.Context) {
	//用来存放从数据库读取出来的用户信息
	data := make([]*dao.UserBasic, 100)
	data = dao.GetUserList()

	c.JSON(200, gin.H {
		"message" : data,
	})
}


//@Summary 登录
//@Tags 用户模块
//@param name query string false "用户名"
//@param password query string false "密码"
//@Success 200 {string} json {"code", "message"}
//@Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	data := dao.UserBasic{}
	//先拿到用户输入的name和password
	//PostForm可以获取客户端通过post请求提交的参数
	name := c.PostForm("name")
	password := c.PostForm("password")

	//然后通过name查找这个人
	user := dao.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H {
			"code" : -1,
			"message" : "该用户不存在",
		})
		return
	}

	//判断数据库的密码和用户输入的密码是否一致
	flag := utils.ValiPassword(password, user.Salt, user.Password)
	//不一致直接返回
	if !flag {
		c.JSON(200, gin.H {
			"code" : -1,
			"message" : "密码错误",
		})
		return
	}

	pwd := utils.MakePassword(password,user.Salt)

	data = dao.FindUserByNameAndPwd(name,pwd)
	c.JSON(200, gin.H {
		"code" : 0,
		"message" : "登录成功",
		"data" :data,
	})
}



//@Summary 新增用户
//@Tags 用户模块
//@param name query string false "用户名"
//@param password query string false "密码"
//@param repassword query string false "确认密码"
//@Success 200 {string} json{"code", "message"}
//@Router /user/createUser [post]
func CreateUser(c *gin.Context) {
	user := dao.UserBasic{}
	name := c.PostForm("name")
	password := c.PostForm("password")
	rePassword := c.PostForm("rePassword")
	salt := fmt.Sprintf("%06d", rand.Int31())

	if password != rePassword {
		c.JSON(-1, gin.H {
			"code" : -1,
			"message" : "两次密码不一致",
		})
		return
	}

	data := dao.FindUserByName(name)
	if name == "" || password == ""  {
		c.JSON(-1,gin.H {
			"code" : -1,
			"message" : "用户名或密码不能为空",
		})
		return
	} else if data.Name != "" {
		c.JSON(-1,gin.H {
			"code" : -1,
			"message" : "该用户已存在",
		})
		return
	}


	user.Name = name
	user.Password = utils.MakePassword(password, salt)
	user.LoginOutTime = time.Now()
	user.LoginTime = time.Now()
	user.HeartbeatTime = time.Now()
	user.Salt = salt
	dao.CreateUser(user)
	c.JSON(200, gin.H {
		"code" : 0,
		"message":"新增用户成功",
	})

}


//@Summary 删除用户
//@Tags 用户模块
//@param id query string false "id"
//@Success 200 {string} json{"code", "message"}
//@Router /user/deleteUser [post]
func DeleteUser(c *gin.Context) {
	//这里使用的是软删除
	user := dao.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	dao.DeleteUser(user)
	c.JSON(200, gin.H {
		"code" : 0,
		"message":"删除用户成功",
	})

}



//@Summary 修改用户
//@Tags 用户模块
//@param id formData string false "id"
//@param name formData string false "name"
//@param password formData string false "password"
//@param phone formData string false "phone"
//@param email formData string false "email"
//@Success 200 {string} json{"code", "message"}
//@Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := dao.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.Password = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Email = c.PostForm("email")

	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(200, gin.H {
			"message" : "修改参数不正确",
		})
	} else {
		dao.UpdateUser(user)
		c.JSON(200, gin.H {
			"message":"修改用户成功",
		})
	}
}
//防止跨域站点伪造请求
var upGrade = websocket.Upgrader{
	//允许跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


func SendMsg(c *gin.Context) {
	//服务器应用程序从http请求处理程序调用Upgrade以获取conn(ws)
	conn, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(conn *websocket.Conn) {
		err = conn.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(conn)
	MsgHandler(conn, c)
}
//发送消息
func MsgHandler(conn *websocket.Conn, c *gin.Context) {
	for {
		msg, err := utils.Subscribe(c, utils.PublishKey)
		if err != nil {
			fmt.Println(" MsgHandler 发送失败", err)
		}

		tm := time.Now().Format("2006-01-02 15:04:05")
		m := fmt.Sprintf("[conn][%s]:%s", tm, msg)
		err = conn.WriteMessage(1, []byte(m))
		if err != nil {
			log.Fatalln(err)
		}
	}
}


func SendUserMsg(c *gin.Context) {
	dao.Chat(c.Writer, c.Request)
}

func SearchFriends(c *gin.Context) {
	id, _ := strconv.Atoi(c.Request.FormValue("userId"))
	//先得到这个user的朋友
	users := dao.SearchFriend(uint(id))
	utils.RespOKList(c.Writer, users, len(users))
}


func AddFriends(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	targetName := c.Request.FormValue("targetName")
	user := dao.FindUserByName(targetName)
	targetId := user.ID
	code, msg := dao.AddFriend(uint(userId), uint(targetId))
	if code ==0 {
		utils.RespOk(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}


func CreateCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	name := c.Request.FormValue("name")
	desc := c.Request.FormValue("desc")
	community := dao.Community{}
	community.OwnerId = uint(ownerId)
	community.Name = name
	community.Desc = desc
	code, msg := dao.CreateCommunity(community)
	if code ==0 {
		utils.RespOk(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

//加载群列表
func LoadCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	data, msg := dao.LoadCommunity(uint(ownerId))
	if len(data) != 0 {
		utils.RespList(c.Writer,0, data, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

//加入群 userId uint, comId uint
func JoinGroups(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	comId := c.Request.FormValue("comId")

	//	name := c.Request.FormValue("name")
	data, msg := dao.JoinGroup(uint(userId), comId)
	if data == 0 {
		utils.RespOk(c.Writer, data, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

func RedisMsg(c *gin.Context) {
	userIdA, _ := strconv.Atoi(c.PostForm("userIdA"))
	userIdB, _ := strconv.Atoi(c.PostForm("userIdB"))
	start, _ := strconv.Atoi(c.PostForm("start"))
	end, _ := strconv.Atoi(c.PostForm("end"))
	isRev, _ := strconv.ParseBool(c.PostForm("isRev"))
	res := dao.RedisMsg(int64(userIdA), int64(userIdB), int64(start), int64(end), isRev)
	utils.RespOKList(c.Writer, "ok", res)
}

func FindByID(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))

	//	name := c.Request.FormValue("name")
	data := dao.FindByID(uint(userId))
	utils.RespOk(c.Writer, data, "ok")
}
