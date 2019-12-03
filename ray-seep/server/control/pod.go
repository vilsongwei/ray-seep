// @File     : pod
// @Author   : Ville
// @Time     : 19-9-26 下午4:40
// node
package control

import (
	jsoniter "github.com/json-iterator/go"
	"ray-seep/ray-seep/common/errs"
	"ray-seep/ray-seep/common/util"
	"ray-seep/ray-seep/conf"
	"ray-seep/ray-seep/proto"
	"ray-seep/ray-seep/server/online"
	"vilgo/vlog"
)

type PodRouterFun func([]byte) (interface{}, error)

// Pod 是一个 代理服务 的管理器连接，包括代理和控制连接
type Pod struct {
	connId   int64
	userId   int64
	appKey   string
	name     string
	secret   string
	domain   string
	out      chan proto.Package
	route    map[int32]PodRouterFun
	userMng  *online.UserManager
	podHd    *PodHandler
	runner   *Runner
	proxyCfg conf.ProxySrv
	protoCfg conf.ProtoSrv
}

func NewPod(id int64, cfg *conf.Server, podHd *PodHandler, out chan proto.Package, runner *Runner) *Pod {
	p := &Pod{
		connId: id,
		podHd:  podHd,
		out:    out,
		runner: runner,
	}
	if cfg.Proto != nil {
		p.protoCfg = *cfg.Proto
	}
	if cfg.Pxy != nil {
		p.proxyCfg = *cfg.Pxy
	}

	p.initRoute()
	return p
}

func (p *Pod) initRoute() {
	p.route = make(map[int32]PodRouterFun)
	p.route[proto.CmdLoginReq] = p.LoginReq
	p.route[proto.CmdCreateHostReq] = p.CreateHostReq
	p.route[proto.CmdRunProxyRsp] = p.RunProxyReq
	p.route[proto.CmdLogoutReq] = p.LogoutReq
}
func (p *Pod) ConnId() int64 {
	return p.connId
}
func (p *Pod) PushMsg(msgPkg *proto.Package) (err error) {
	p.out <- *msgPkg
	return
}

func (p *Pod) OnMessage(cmd int32, body []byte) ([]byte, error) {
	if rt, ok := p.route[cmd]; ok {
		rsp, err := rt(body)
		if err != nil {
			return nil, err
		}
		rspBody, err := jsoniter.Marshal(rsp)
		if err != nil {
			return nil, err
		}
		return rspBody, nil
	}
	return nil, errs.ErrNoCmdRouterNot
}

func (p *Pod) LoginReq(req []byte) (interface{}, error) {
	reqLogin := proto.LoginReq{}
	resp := &proto.LoginRsp{Id: p.connId, Token: util.RandToken()}

	if err := jsoniter.Unmarshal(req, &reqLogin); err != nil {
		vlog.ERROR("[%d] login Unmarshal error:%s", p.connId, err.Error())
		return nil, err
	}

	vlog.INFO("[%d] login request userId=%d ", p.connId, reqLogin.UserId)
	secret, err := p.podHd.OnLogin(p.connId, reqLogin.UserId, reqLogin.AppKey, resp.Token)
	if err != nil {
		vlog.ERROR("[%d] login store token error:%s", p.connId, err.Error())
		return nil, err
	}
	p.appKey = reqLogin.AppKey
	p.name = reqLogin.Name
	p.secret = secret
	return resp, nil
}

func (p *Pod) CreateHostReq(req []byte) (rsp interface{}, err error) {
	reqObj := &proto.CreateHostReq{}
	if err = jsoniter.Unmarshal(req, reqObj); err != nil {
		vlog.ERROR("")
		return
	}

	addr := ":4900"

	join := JoinItem{
		Name:   reqObj.SubDomain,
		ConnId: p.connId,
		Addr:   addr,
		Err:    make(chan error),
	}

	p.runner.Join() <- join
	if err = <-join.Err; err != nil {
		vlog.ERROR("[%d] http join error %s", p.connId, err.Error())
		return
	}

	rspObj := proto.CreateHostRsp{
		ProxyPort:  p.proxyCfg.Port,
		HttpDomain: reqObj.SubDomain + "." + p.protoCfg.Domain + addr,
	}
	return rspObj, nil
}

// NoticeRunProxy 通知用户启动代理服务服务
func (p *Pod) NoticeRunProxy() {
	notice := &proto.Package{
		Cmd:  proto.CmdNoticeRunProxy,
		Body: nil,
	}

	if err := p.PushMsg(notice); err != nil {
		vlog.ERROR("[%d] notice run proxy error %s", p.connId, err.Error())
	}
}

func (p *Pod) RunProxyReq(req []byte) (rsp interface{}, err error) {
	reqObj := &proto.RunProxyReq{}
	if err = jsoniter.Unmarshal(req, reqObj); err != nil {
		vlog.ERROR("run proxy request error %s", err.Error())
		return
	}
	return proto.RunProxyRsp{}, nil
}

func (p *Pod) LogoutReq(req []byte) (rsp interface{}, err error) {
	reqObj := make(map[string]interface{})
	if err = jsoniter.Unmarshal(req, &reqObj); err != nil {
		return
	}
	if reqObj["IsClean"].(bool) {
		p.runner.Leave()
	}
	vlog.INFO("能否彻底关掉%v", reqObj)
	return nil, nil
}
