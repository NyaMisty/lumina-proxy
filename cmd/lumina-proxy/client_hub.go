package main

import (
	"context"
	"fmt"
	"log"

	"github.com/palantir/stacktrace"
	"github.com/zhangyoufu/lumina"
)

type ClientSessionHub struct {
	Clients  []*lumina.Client
	Sessions []*lumina.ClientSession
}

func DialClients(clients []*lumina.Client, ctx context.Context, logger *log.Logger, version int32, interpreter lumina.Interpreter) (ret *ClientSessionHub, err error) {
	ret = &ClientSessionHub{}
	ret.Sessions = make([]*lumina.ClientSession, len(clients))
	hasSucc := false
	for i, client := range clients {
		sess, _err := client.Dial(ctx, logger, version, interpreter)
		if _err != nil {
			if _err != context.Canceled {
				err = stacktrace.Propagate(_err, "unable to create upstream session for "+client.Dialer.Info())
			} else {
				err = fmt.Errorf("dial %s error: %w", client.Dialer.Info(), _err)
			}
			logger.Print(err)
			ret.Sessions[i] = nil
		}
		ret.Sessions[i] = sess
		hasSucc = true
	}
	err = nil
	if !hasSucc {
		err = fmt.Errorf("all upstream connection failed")
	}
	return
}

func (h *ClientSessionHub) Request(ctx context.Context, req lumina.Request) (rsps []lumina.Packet, err error) {
	rsps = make([]lumina.Packet, 0, len(h.Sessions))
	for i, session := range h.Sessions {
		if session == nil {
			continue
		}
		rsp, _err := session.Request(ctx, req)
		if _err != nil {
			err = fmt.Errorf("error during %s request: %w", h.Clients[i].Dialer.Info(), _err)
			continue
			//return
		}
		rsps = append(rsps, rsp)
	}
	return
}
