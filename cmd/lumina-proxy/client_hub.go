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
	for i, client := range clients {
		sess, _err := client.Dial(ctx, logger, version, interpreter)
		if _err != nil {
			if _err != context.Canceled {
				err = stacktrace.Propagate(_err, "unable to create upstream session for "+client.Dialer.Info())
			} else {
				err = fmt.Errorf("dial %s error: %w", client.Dialer.Info(), _err)
			}
			return
		}
		ret.Sessions[i] = sess
	}
	return
}

func (h *ClientSessionHub) Request(ctx context.Context, req lumina.Request) (rsps []lumina.Packet, err error) {
	rsps = make([]lumina.Packet, len(h.Sessions))
	for i, session := range h.Sessions {
		rsp, _err := session.Request(ctx, req)
		if _err != nil {
			err = fmt.Errorf("error during %s request: %w", h.Clients[i].Dialer.Info(), _err)
			return
		}
		rsps[i] = rsp
	}
	return
}
