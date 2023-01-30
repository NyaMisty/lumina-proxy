package main

import (
	"context"
	"github.com/palantir/stacktrace"
	"github.com/zhangyoufu/lumina"
)

// 'W' for Windows
// 'M' for macOS
// 'L' for Linux
var licOS byte

type Proxy struct {
	lumina.Server
}

func NewProxy(licKey lumina.LicenseKey, licId lumina.LicenseId) (proxy *Proxy) {
	client := &lumina.Client{
		LicenseKey: licKey,
		LicenseId:  licId,
	}
	return NewProxyEx([]*lumina.Client{client})
}

func NewProxyEx(clients []*lumina.Client) (proxy *Proxy) {
	handler := &proxyHandler{}
	proxy = &Proxy{}
	proxy.Handler = handler
	proxy.OnHELO = func(ctx context.Context) (newctx context.Context, err error) {
		sessions, err := DialClients(clients, ctx, lumina.GetLogger(ctx), lumina.GetProtocolVersion(ctx), handler)
		if err != nil {
			if err != context.Canceled {
				err = stacktrace.Propagate(err, "unable to create upstream session")
			}
			return
		}
		newctx = setUpstream(ctx, sessions)
		return
	}
	return
}

type proxyHandler struct{}

// Currently, we only allow pulling/pushing.
func (*proxyHandler) AcceptRequest(t lumina.PacketType) bool {
	switch t {
	case lumina.PKT_PULL_MD:
		return true
	case lumina.PKT_PUSH_MD:
		return true
	case lumina.PKT_DECOMPILE:
		return true
	default:
		return false
	}
}

func (*proxyHandler) GetPacketOfType(t lumina.PacketType) lumina.Packet {
	switch t {
	case lumina.PKT_PULL_MD:
		return &lumina.PullMdPacket{}
	case lumina.PKT_PULL_MD_RESULT:
		return &lumina.PullMdResultPacket{}
	case lumina.PKT_PUSH_MD:
		return &lumina.PushMdPacket{}
	case lumina.PKT_PUSH_MD_RESULT:
		return &lumina.PushMdResultPacket{}
	case lumina.PKT_DECOMPILE:
		return &lumina.DecompilePacket{}
	case lumina.PKT_DECOMPILE_RESULT:
		return &lumina.DecompileResultPacket{}
	default:
		return nil
	}
}

// Pump between client and upstream server. (half-duplex)
func (*proxyHandler) ServeRequest(ctx context.Context, req lumina.Request) (finalRsp lumina.Packet, err error) {
	if pkt, ok := req.(*lumina.PushMdPacket); ok {
		pkt.AnonymizeFields(ctx)
	}
	rsps, _err := getUpstream(ctx).Request(ctx, req)
	if _err != nil {
		if len(rsps) == 0 {
			err = _err
			return
		}
		lumina.GetLogger(ctx).Printf("partial error during request, we'll ignore it!")
	}

	switch rsps[0].(type) {
	case *lumina.PullMdResultPacket:
		firstRsp := rsps[0].(*lumina.PullMdResultPacket)
		rsp := &lumina.PullMdResultPacket{}
		rsp.Codes = make([]lumina.OpResult, len(firstRsp.Codes))
		rsp.Results = make([]lumina.FuncInfoAndFrequency, 0)
		for i, _ := range rsp.Codes {
			rsp.Codes[i] = lumina.PDRES_ERROR
		}
		for _, curRspRaw := range rsps {
			if curRsp, ok := curRspRaw.(*lumina.PullMdResultPacket); ok {
				for i, code := range curRsp.Codes {
					if rsp.Codes[i] == lumina.PDRES_OK {
						continue
					}
					rsp.Codes[i] = code
				}
				rsp.Results = append(rsp.Results, curRsp.Results...)
			}
		}
		finalRsp = rsp
	case *lumina.PushMdResultPacket:
		firstRsp := rsps[0].(*lumina.PushMdResultPacket)
		rsp := &lumina.PushMdResultPacket{}
		rsp.Codes = make([]lumina.OpResult, len(firstRsp.Codes))
		for i, _ := range rsp.Codes {
			rsp.Codes[i] = lumina.PDRES_ERROR
		}
		for _, curRspRaw := range rsps {
			if curRsp, ok := curRspRaw.(*lumina.PullMdResultPacket); ok {
				for i, code := range curRsp.Codes {
					if rsp.Codes[i] == lumina.PDRES_OK {
						continue
					}
					rsp.Codes[i] = code
				}
			}
		}
		finalRsp = rsp
	}
	// if err != nil {
	//     lumina.GetConn(ctx).Close()
	// }
	return
}
