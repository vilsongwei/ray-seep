package control

import (
	"ray-seep/ray-seep/common/errs"
	"ray-seep/ray-seep/databus"
	"ray-seep/ray-seep/model"
)

type PodHandler struct {
	db databus.BaseDao
}

func NewPodHandler(db databus.BaseDao) *PodHandler {
	return &PodHandler{db: db}
}

func (sel *PodHandler) OnLogin(connId, userId int64, user string, appKey string, token string) (loginDao *model.UserLoginDao, err error) {
	loginDao, err = sel.db.UserLogin(connId, userId, user, appKey, token)
	if err != nil {
		return
	}
	if loginDao.HttpPort == "" {
		return nil, errs.ErrHttpPortIsInValid
	}
	return
}

func (sel *PodHandler) OnLogout(name string, id int64) error {
	return nil
}

// 创建主机判断是否登录
func (sel *PodHandler) OnCreateHost(id int64, token string) error {
	return nil
}
