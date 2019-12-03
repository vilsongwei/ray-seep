package control

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"ray-seep/ray-seep/client/proxy"
	"ray-seep/ray-seep/conf"
	"ray-seep/ray-seep/proto"
	"time"
	"vilgo/vlog"
)

type ClientControlHandler struct {
	connId       int64
	userId       int64
	token        string
	secret       string
	appKey       string
	domain       string
	name         string // 子域名
	pxyHost      string
	pxyPort      int64
	push         ResponsePush
	cliPxy       *proxy.ClientProxy
	cliPxyStopCh chan int
}

func NewClientControlHandler(cfg *conf.Client) *ClientControlHandler {
	stopCh := make(chan int)
	return &ClientControlHandler{
		name:         cfg.Control.Name,
		cliPxyStopCh: stopCh,
		cliPxy:       proxy.NewClientProxy(stopCh, cfg),
		appKey:       cfg.Control.AppKey,
		userId:       cfg.Control.UserId,
		secret:       cfg.Control.Secret,
		pxyHost:      cfg.Pxy.Address,
	}
}

func (c *ClientControlHandler) Ping() {
	go func() {
		tm := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-tm.C:
				if err := c.push.PushEvent(proto.CmdPing, nil); err != nil {
					return
				}
			}
		}
	}()
	return
}

func (c *ClientControlHandler) Pong(req *proto.Package) (err error) {
	//vlog.INFO("server message  pong [%d]", req.Cmd)
	return
}

// 登录服务器
func (c *ClientControlHandler) Login(push ResponsePush) (err error) {

	dt, err := jsoniter.Marshal(&proto.LoginReq{UserId: c.userId, Name: c.name, AppKey: c.appKey})
	if err != nil {
		vlog.ERROR("push event json marshal error  %s", err.Error())
		return err
	}
	c.push = push
	return c.push.PushEvent(proto.CmdLoginReq, dt)
}

func (c *ClientControlHandler) LoginRsp(req *proto.Package) (err error) {
	rsp := &proto.LoginRsp{}
	if err := jsoniter.Unmarshal(req.Body, rsp); err != nil {
		return err
	}
	vlog.INFO("login success")
	c.connId = rsp.Id
	c.token = rsp.Token
	c.Ping()
	return c.CreateHostReq()
}

//  CreateHostReq 创建服务主机
func (c *ClientControlHandler) CreateHostReq() error {
	reqData, err := jsoniter.Marshal(proto.CreateHostReq{SubDomain: c.name})
	if err != nil {
		return err
	}
	return c.push.PushEvent(proto.CmdCreateHostReq, reqData)
}

// CreateHostRsp 创建服务主机返回
func (c *ClientControlHandler) CreateHostRsp(req *proto.Package) (err error) {
	ctInfo := &proto.CreateHostRsp{}
	if err = jsoniter.Unmarshal(req.Body, ctInfo); err != nil {
		vlog.ERROR("create host response json un parse error %s", err.Error())
		return
	}
	vlog.INFO("")
	vlog.INFO("\t---------------------create host success-----------------------")
	vlog.INFO("\t\t     user_id : %d ", c.userId)
	vlog.INFO("\t\t      secret : %s ", c.secret)
	vlog.INFO("\t\t     app_key : %s ", c.appKey)
	vlog.INFO("\t\t     conn_id : %d ", c.connId)
	vlog.INFO("\t\t       token : %s ", c.token)
	vlog.INFO("\t\t   http host : %s ", ctInfo.HttpDomain)
	vlog.INFO("\t---------------------------------------------------------------")
	c.domain = ctInfo.HttpDomain
	c.pxyPort = ctInfo.ProxyPort
	// 收到创建主机的返回信息就可 运行代理了
	return c.RunProxyReq()
}

// NoticeRunProxy 通知创建代理
func (c *ClientControlHandler) NoticeRunProxy(req *proto.Package) error {
	//vlog.INFO("收到 [NoticeRunProxy]Cmd:%d Body:%s", req.Cmd, string(req.Body))
	return c.RunProxyReq()
}

func (c *ClientControlHandler) RunProxyReq() (err error) {
	return c.cliPxy.RunProxy(c.connId, c.token, c.name, fmt.Sprintf("%s:%d", c.pxyHost, c.pxyPort))
}

func (c *ClientControlHandler) RunProxyRsp(req *proto.Package) (err error) {
	//vlog.INFO("收到 [RunProxyRsp]Cmd:%d Body:%s", req.Cmd, string(req.Body))
	return nil
}

func (c *ClientControlHandler) LogoutRsp(req *proto.Package) (err error) {
	vlog.INFO("disconnect cid:%d", c.connId)
	return nil
}