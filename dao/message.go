package dao

import (
	"IMProject/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"gopkg.in/fatih/set.v0"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Message struct {
	gorm.Model
	UserId int64 //发送者
	TargetId int64 //接收者
	Type int // 发送类型, 群聊，私聊，广播
	Media int //消息类型 文字 图片 音频
	content string //消息内容
	CreateTime uint64 //创建时间
	ReadTime   uint64 //读取时间
	Pic string
	Url string
	Desc string
	Amount int
}

func (table *Message) TableName() string {
	return "message"
}

type Node struct {
	Conn          *websocket.Conn //连接
	Addr          string          //客户端地址
	FirstTime     uint64          //首次连接时间
	HeartbeatTime uint64          //心跳时间
	LoginTime     uint64          //登录时间
	DataQueue     chan []byte     //消息
	GroupSets     set.Interface   //好友 / 群
}

//首先前端将消息以及消息信息发送给发送人的websocket，
//协程从发送人的websocket取出以后，然后将信息发送给接收人的管道
//协程将信息从接收人的管道中取出再发送给接收人的websocket


//映射关系
var clientMap map[int64] *Node = make(map[int64] *Node, 0)

//读写锁
var rwLocker sync.RWMutex

func Chat(writer http.ResponseWriter, request *http.Request) {
	//1.获取参数并检验token等合法性
	//token := query.Get("token")
	query := request.URL.Query()
	//从访问路径中拿到userId
	id := query.Get("userId")
	userId, _ := strconv.ParseInt(id, 10, 64)
	//msgType := query.Get("type")
	//targetId := query.Get("targetId")
	//context := query.Get("context")
	isValide := true
	//完成握手，升级为websocket长连接，使用conn发送和接收消息
	conn, err := (&websocket.Upgrader{
		//允许跨域
		CheckOrigin : func(r *http.Request) bool {
			return isValide
		},
	}).Upgrade(writer, request, nil)


	if err != nil {
		fmt.Println(err)
		return
	}

	//2.获取conn
	currentTime := uint64(time.Now().Unix())
	//生成node，一个websocket连接对应一个队列
	node := &Node {
		Conn : conn,
		Addr:          conn.RemoteAddr().String(), //客户端地址
		HeartbeatTime: currentTime,                //心跳时间
		LoginTime:     currentTime,                //登录时间
		DataQueue : make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
	}
	//用户关系
	//userId跟Node绑定，并加锁，以防并发问题
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	//发送
	go sendProc(node)
	//接收
	go recvProc(node)

	SetUserOnlineInfo("online_"+id, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)
	sendMsg(userId,[]byte("欢迎进入聊天室"))
}


//从队列中取数据并放进websocket里面
func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			fmt.Println("[ws]sendMsg >>>>>>> msg:",string(data))
			//将消息发到websocket
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

//从websocket读数据并把数据放进队列里面
func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := Message{}
		err = json.Unmarshal(data, &msg)
		if err != nil {
			fmt.Println(err)
		}
		//心跳检测 msg.Media == -1 || msg.Type == 3
		if msg.Type == 3 {
			currentTime := uint64(time.Now().Unix())
			node.Heartbeat(currentTime)
		} else {
			dispatch(data)
			broadMsg(data) //todo 将消息广播到局域网
			fmt.Println("[ws] recvProc <<<<< ", string(data))
		}

	}
}


var updsendChan chan []byte = make(chan []byte, 1024)
func broadMsg(data []byte) {
	updsendChan<- data
}


func init() {
	go udpSendProc()
	go udpRecvProc()
	fmt.Println("init goroutine ")
}


//完成upd数据发送协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP : net.IPv4(192 ,168 ,0 ,255),
		Port: viper.GetInt("port.udp"),
	})

	defer con.Close()
	if err != nil {
		fmt.Println(err)
	}
	for {
		select {
		case data := <-updsendChan:
			fmt.Println("udpSendProc >>>>>>> data :", string(data))
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}

}

//完成upd数据接收协程

func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: viper.GetInt("port.udp"),
	})
	if err != nil {
		fmt.Println(err)
	}
	defer con.Close()
	for {
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("udpRecvProc data:",string(buf[0:n]))
		dispatch(buf[0:n])
	}
}


func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch msg.Type {
	case 1 : //私信
		fmt.Println("dispatch data:", string(data))
		sendMsg(msg.TargetId, data)
	case 2: //群发,这个targetId指的是群的ID，相当于是找到群里人所有人的ID，然后再给他私发
		sendGroupMsg(msg.TargetId, data)
	//case 3: //广播
		//sendAllMsg()
	}
}

func sendGroupMsg(targetId int64, msg []byte) {
	fmt.Println("开始群发消息")
	//找到在这个群的其他人的id
	userIds := SearchUserByGroupId(uint(targetId))
	for i := 0; i < len(userIds); i++ {
		sendMsg(int64(userIds[i]), msg)
	}
}

//发送消息，拿到发送者id，拿到接收者id，从redis中看接收者是否在线，如果不在线就没法发
//如果在线，就将消息发送给发送者的那个队列里面，等待select去拿
//然后将消息放进redis里面存上
func sendMsg(userId int64, msg []byte) {
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()
	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	//拿到接收者id
	targetIdStr := strconv.Itoa(int(userId))
	//拿到发送者id
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = uint64(time.Now().Unix())
	//看redis中接收者是否在线
	r, err := utils.Red.Get(ctx, "online_"+userIdStr).Result()
	if err != nil {
		fmt.Println(err)
	}
	if r != "" {
		if ok {
			//如果没问题，就往消息队列里面发送那个消息
			fmt.Println("sendMsg >>> userID: ", userId, "  msg:", string(msg))
			node.DataQueue <- msg
		}
	}
	var key string
	//将redis中key的名字都存成msg_小ID_大ID
	if userId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}
	//返回有序集key中指定区间内的成员，
	//其中成员的位置按score值递减来排列，具有相同的score值的成员按字典序的逆序排列
	res, err := utils.Red.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	score := float64(cap(res)) + 1
	//将一个或多个成员元素及其分数值加入到有序集中
	ress, e := utils.Red.ZAdd(ctx, key, &redis.Z{score, msg}).Result() //jsonMsg
	//res, e := utils.Red.Do(ctx, "zadd", key, 1, jsonMsg).Result() //备用 后续拓展 记录完整msg
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(ress)
}

func JoinGroup(userId uint, comId string) (int, string) {
	contact := Contact{}
	//ownerId是本人的ID，targetId是群的ID
	contact.OwnerId = userId
	contact.Type = 2
	community := Community{}

	utils.DB.Where("id=? or name=?", comId, comId).Find(&community)
	if community.Name == "" {
		return -1, "没有找到群"
	}
	utils.DB.Where("owner_id=? and target_id=? and type =2 ", userId, community.ID).Find(&contact)
	if !contact.CreatedAt.IsZero() {
		return -1, "已加过此群"
	} else {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功"
	}
}

//获取缓存里面的消息
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {
	rwLocker.RLock()
	rwLocker.RUnlock()
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}

	var rels []string
	var err error
	if isRev {
		rels, err = utils.Red.ZRange(ctx, key, start, end).Result()
	} else {
		rels, err = utils.Red.ZRevRange(ctx, key, start, end).Result()
	}
	if err != nil {
		fmt.Println(err) 
	}
	return rels
}

//更新用户心跳
func (node *Node) Heartbeat(currentTime uint64) {
	node.HeartbeatTime = currentTime
	return
}

//需要重写此方法才能完整的msg转byte[]
func (msg Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(msg)
}

//清理超时连接
func CleanConnection(param interface{}) (result bool) {
	result = true
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("cleanConnection err", r)
		}
	}()
	//fmt.Println("定时任务,清理超时连接 ", param)
	//node.IsHeartbeatTimeOut()
	currentTime := uint64(time.Now().Unix())
	for i := range clientMap {
		node := clientMap[i]
		if node.IsHeartbeatTimeOut(currentTime) {
			fmt.Println("心跳超时..... 关闭连接：", node)
			node.Conn.Close()
		}
	}
	return result
}

//用户心跳是否超时
func (node *Node) IsHeartbeatTimeOut(currentTime uint64) (timeout bool) {
	if node.HeartbeatTime+viper.GetUint64("timeout.HeartbeatMaxTime") <= currentTime {
		fmt.Println("心跳超时。。。自动下线", node)
		timeout = true
	}
	return
}












