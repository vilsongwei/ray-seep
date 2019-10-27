package proxy

import (
	"encoding/json"
	"fmt"
	"ray-seep/ray-seep/common/conn"
	"ray-seep/ray-seep/common/pkg"
	"ray-seep/ray-seep/conf"
	"ray-seep/ray-seep/mng"
	"time"
	"vilgo/vlog"
)

type IRegister interface {
	Register(domain string, cc conn.Conn) error
}

type ProxyServer struct {
	addr      string
	proxyConn chan conn.Conn
	register  IRegister //
}

func NewProxyServer(c *conf.ProxySrv, reg IRegister) *ProxyServer {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	return &ProxyServer{
		addr:      addr,
		proxyConn: make(chan conn.Conn),
		register:  reg,
	}
}

func (s *ProxyServer) Start() {
	ls, err := conn.Listen(s.addr)
	if err != nil {
		return
	}
	vlog.INFO("ProxyServer start [%s]", s.addr)
	for c := range ls.Conn {
		go s.dealConn(c)
	}
}

func (s *ProxyServer) dealConn(cn conn.Conn) {
	defer func() {
		if err := recover(); err != nil {
			vlog.DEBUG("")
			return
		}
	}()
	_ = cn.SetDeadline(time.Now().Add(time.Second * 15))
	tr := mng.NewMsgTransfer(cn)
	var regProxy pkg.Package
	if err := tr.RecvMsg(&regProxy); err != nil {
		vlog.ERROR("receive message error %s", err.Error())
		_ = cn.Close()
		return
	}

	if regProxy.Cmd != pkg.CmdRegisterProxyReq {
		vlog.ERROR("proxy cmd is error %d", regProxy.Cmd)
		_ = cn.Close()
		return
	}

	regData := pkg.RegisterProxyReq{}
	if err := json.Unmarshal(regProxy.Body, &regData); err != nil {
		vlog.ERROR("parse register proxy request data fail %s , data is %s ", err.Error(), string(regProxy.Body))
		_ = cn.Close()
		return
	}
	// 把代理连接都注册到注册器里面
	if err := s.register.Register(regData.SubDomain, cn); err != nil {
		vlog.ERROR("%s proxy is registered fail %s", cn.RemoteAddr().String(), err.Error())
		_ = cn.Close()
		return
	}
}