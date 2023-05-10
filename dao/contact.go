package dao

import (
	"IMProject/utils"
	"fmt"
	"gorm.io/gorm"
)

//人员关系
type Contact struct {
	gorm.Model
	OwnerId uint//谁的关系
	TargetId uint //朋友ID
	Type int//对应的类型 1好友 2 群
	Desc string

}

func (table *Contact) TableName() string {
	return "contact"
}


func SearchFriend(userId uint) ([]UserBasic) {
	contacts := make([] Contact, 0)
	objIDs := make([]uint64, 0)
	utils.DB.Where("owner_id = ? and type =1", userId).Find(&contacts)
	for _, v := range contacts {
		fmt.Println("=====================", v)
		//获取到好友的ID
		objIDs = append(objIDs, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0)
	//查询好友的信息
	utils.DB.Where("id in ?",objIDs).Find(&users)
	return users
}

//userId是本人ID，targetId是好友ID
func AddFriend(userId uint, targetId uint) (int, string){
	user := UserBasic{}
	if targetId !=0 {
		user = FindByID(targetId)
		if user.Salt != "" {
			//如果本人ID等于根据targetID,表示添加的是自己
			if userId == targetId {
				return -1,"不能添加自己"
			}
			contact0 := Contact{}
			//是否添加过
			utils.DB.Where("owner_id =? and target_id = ? and type =1", userId, targetId).Find(&contact0)
			if contact0.ID != 0 {
				return -1, "不能重复添加"
			}

			tx := utils.DB.Begin()
			//事务一旦开始,出现异常就回滚
			defer func () {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()

			contact := Contact{}
			contact.OwnerId = userId
			contact.TargetId = targetId
			contact.Type = 1
			if err := utils.DB.Create(&contact).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}
			contact1 := Contact{}
			contact1.OwnerId = targetId
			contact1.TargetId = userId
			contact1.Type = 1
			utils.DB.Create(&contact1)

			if err := utils.DB.Create(&contact1).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}
			tx.Commit()
			return 0, "添加好友成功"
		}
		return -1, "没有找到此用户"
	}
	return -1, "好友ID不能为空"
}

func SearchUserByGroupId(communityId uint) []uint {
	contacts := make([]Contact, 0)
	objIds := make([]uint, 0)
	utils.DB.Where("target_id = ? and type=2", communityId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, uint(v.OwnerId))
	}
	return objIds
}













