// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/cdp"
	jsonv2 "github.com/chromedp/cdproto/cdp/jsonv2"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/remote"
	"go.gearno.de/kit/httpclient"
)

const childCallTimeout = 5 * time.Second

type devtools struct {
	conn *remote.Conn

	mu    sync.Mutex
	next  int64
	wait  map[int64]chan *cdproto.Message
	armed map[target.ID]struct{}
}

func (s *scan) watchChildren(ctx context.Context) error {
	if s.in.BrowserURL == "" {
		if s.in.Logger != nil {
			s.in.Logger.InfoCtx(ctx, "cookie scan is not instrumenting child targets")
		}

		return nil
	}

	info := chromedp.FromContext(ctx)
	if info == nil || info.Target == nil {
		return fmt.Errorf("cannot attach child targets: browser target is missing")
	}

	s.browserContext = string(info.BrowserContextID)

	wsURL, err := browserWebsocket(ctx, s.in.BrowserURL)
	if err != nil {
		return fmt.Errorf("cannot resolve browser websocket: %w", err)
	}

	conn, err := remote.DialContext(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("cannot attach child targets: %w", err)
	}

	d := &devtools{
		conn:  conn,
		wait:  map[int64]chan *cdproto.Message{},
		armed: map[target.ID]struct{}{},
	}

	go d.read(s)

	go func() {
		<-s.scanCtx.Done()

		_ = conn.Close()
	}()

	if err := d.call(ctx, "", target.AutoAttachRelated.Method, target.AutoAttachRelatedParams{
		TargetID:               info.Target.TargetID,
		WaitForDebuggerOnStart: true,
		Filter:                 childTargetFilter(),
	}); err != nil {
		_ = conn.Close()

		return fmt.Errorf("cannot watch child targets: %w", err)
	}

	return nil
}

func (d *devtools) read(s *scan) {
	for {
		msg := &cdproto.Message{}
		if err := d.conn.Read(s.scanCtx, msg); err != nil {
			if s.scanCtx.Err() == nil {
				s.noteStreamErr(err)
			}

			return
		}

		if msg.ID != 0 {
			d.deliver(msg)

			continue
		}

		ev, err := cdproto.UnmarshalMessage(msg)
		if err != nil {
			continue
		}

		d.dispatch(s, msg.SessionID, ev)
	}
}

func (d *devtools) deliver(msg *cdproto.Message) {
	d.mu.Lock()
	ch := d.wait[msg.ID]
	d.mu.Unlock()

	if ch == nil {
		return
	}

	select {
	case ch <- msg:
	default:
	}
}

func (d *devtools) dispatch(s *scan, session target.SessionID, ev any) {
	switch ev := ev.(type) {
	case *target.EventAttachedToTarget:
		go s.onAttached(d, ev)
	case *target.EventDetachedFromTarget:
		s.rec.dropSession(string(ev.SessionID))
	default:
		if session == "" {
			return
		}

		s.onChildEvent(d, session, ev)
	}
}

func (s *scan) onAttached(d *devtools, ev *target.EventAttachedToTarget) {
	info := ev.TargetInfo
	if info == nil || !s.wantsChild(info) {
		s.releaseChild(d, ev.SessionID)

		return
	}

	if !d.mark(info.TargetID) {
		s.releaseChild(d, ev.SessionID)

		return
	}

	ctx, cancel := context.WithTimeout(s.scanCtx, childCallTimeout)
	defer cancel()

	if err := d.call(ctx, ev.SessionID, network.Enable.Method, network.EnableParams{}); err != nil {
		s.warn(ctx, "cannot enable child network events", info.URL, err)
	}

	if err := d.call(ctx, ev.SessionID, network.SetBypassServiceWorker.Method, network.SetBypassServiceWorkerParams{Bypass: true}); err != nil {
		s.warn(ctx, "cannot bypass child service workers", info.URL, err)
	}

	if err := d.call(ctx, ev.SessionID, fetch.Enable.Method, fetch.EnableParams{}); err != nil {
		s.warn(ctx, "cannot intercept child requests", info.URL, err)
	}

	if info.Type == "iframe" {
		if err := d.call(ctx, ev.SessionID, runtime.AddBinding.Method, runtime.AddBindingParams{Name: s.binding}); err != nil {
			s.warn(ctx, "cannot add child cookie binding", info.URL, err)
		}

		if err := d.call(ctx, ev.SessionID, page.AddScriptToEvaluateOnNewDocument.Method, page.AddScriptToEvaluateOnNewDocumentParams{
			Source: cookieHookSource(s.binding),
		}); err != nil {
			s.warn(ctx, "cannot install child cookie hook", info.URL, err)
		}
	}

	if err := d.call(ctx, "", target.AutoAttachRelated.Method, target.AutoAttachRelatedParams{
		TargetID:               info.TargetID,
		WaitForDebuggerOnStart: true,
		Filter:                 childTargetFilter(),
	}); err != nil {
		s.warn(ctx, "cannot watch nested targets", info.URL, err)
	}

	s.releaseChild(d, ev.SessionID)
}

func (s *scan) wantsChild(info *target.Info) bool {
	if !childTargetType(info.Type) {
		return false
	}

	if s.browserContext != "" && info.BrowserContextID != "" && string(info.BrowserContextID) != s.browserContext {
		return false
	}

	return true
}

func (s *scan) releaseChild(d *devtools, session target.SessionID) {
	ctx, cancel := context.WithTimeout(s.scanCtx, childCallTimeout)
	defer cancel()

	_ = d.call(ctx, session, runtime.RunIfWaitingForDebugger.Method, cdp.Empty{})
}

func (s *scan) onChildEvent(d *devtools, session target.SessionID, ev any) {
	if paused, ok := ev.(*fetch.EventRequestPaused); ok {
		go s.answerChild(d, session, paused)

		return
	}

	s.rec.handleSession(string(session), ev)
}

func (s *scan) answerChild(d *devtools, session target.SessionID, ev *fetch.EventRequestPaused) {
	block := ev.Request == nil
	if ev.Request != nil {
		block = s.policy.block(
			s.scanCtx,
			ev.Request.Method,
			ev.Request.URL,
			string(ev.ResourceType),
			string(ev.FrameID),
		)
	}

	ctx, cancel := context.WithTimeout(s.scanCtx, childCallTimeout)
	defer cancel()

	if block {
		_ = d.call(ctx, session, fetch.FailRequest.Method, fetch.FailRequestParams{
			RequestID:   ev.RequestID,
			ErrorReason: network.ErrorReasonBlockedByClient,
		})

		return
	}

	_ = d.call(ctx, session, fetch.ContinueRequest.Method, fetch.ContinueRequestParams{
		RequestID: ev.RequestID,
	})
}

func (d *devtools) mark(id target.ID) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.armed[id]; ok {
		return false
	}

	d.armed[id] = struct{}{}

	return true
}

func (d *devtools) call(ctx context.Context, session target.SessionID, method string, params any) error {
	id := atomic.AddInt64(&d.next, 1)

	var raw jsonv2.Value

	if params != nil {
		buf, err := jsonv2.Marshal(params, chromedp.DefaultMarshalOptions)
		if err != nil {
			return fmt.Errorf("cannot encode browser command: %w", err)
		}

		raw = buf
	}

	ch := make(chan *cdproto.Message, 1)

	d.mu.Lock()
	d.wait[id] = ch
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		delete(d.wait, id)
		d.mu.Unlock()
	}()

	if err := d.conn.Write(ctx, &cdproto.Message{
		ID:        id,
		SessionID: session,
		Method:    cdproto.MethodType(method),
		Params:    raw,
	}); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case msg := <-ch:
		if msg.Error != nil {
			return msg.Error
		}

		return nil
	}
}

func childTargetFilter() target.Filter {
	return target.Filter{
		{Type: "iframe"},
		{Type: "worker"},
		{Type: "shared_worker"},
		{Type: "service_worker"},
		{Exclude: true},
	}
}

func childTargetType(typ string) bool {
	switch typ {
	case "iframe", "worker", "shared_worker", "service_worker":
		return true
	default:
		return false
	}
}

// browserWebsocket resolves the Chrome debugger address. The host is the
// browser this scan already drives, so the lookup is not customer traffic.
func browserWebsocket(ctx context.Context, raw string) (string, error) {
	if strings.Contains(raw, "/devtools/browser/") {
		return raw, nil
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("cannot parse browser URL: %w", err)
	}

	switch parsed.Scheme {
	case "ws":
		parsed.Scheme = "http"
	case "wss":
		parsed.Scheme = "https"
	case "http", "https":
	default:
		return "", fmt.Errorf("cannot parse browser URL: unsupported scheme %q", parsed.Scheme)
	}

	parsed.Path = "/json/version"
	parsed.RawQuery = ""
	parsed.Fragment = ""

	reqCtx, cancel := context.WithTimeout(ctx, childCallTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", fmt.Errorf("cannot request browser version: %w", err)
	}

	resp, err := httpclient.DefaultClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot request browser version: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	var version struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return "", fmt.Errorf("cannot read browser version: %w", err)
	}

	if version.WebSocketDebuggerURL == "" {
		return "", fmt.Errorf("cannot find browser websocket")
	}

	return version.WebSocketDebuggerURL, nil
}
