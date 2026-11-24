package app

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// The issue is written with what was typed: its first line the title, all
// of it and the versions the body. Nothing typed, the templates open; too
// long for an address, the body is left to paste.
func TestFeedbackIssueURL(t *testing.T) {
	link, long := feedbackIssueURL("  桌面歌词偶尔会闪\n切歌的时候  ")
	if long {
		t.Fatal("a short message is long")
	}
	u, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "github.com" || u.Path != "/nya-owo/cloudy/issues/new" {
		t.Errorf("link %q", link)
	}
	if got := u.Query().Get("title"); got != "桌面歌词偶尔会闪" {
		t.Errorf("title %q", got)
	}
	if body := u.Query().Get("body"); !strings.HasPrefix(body, "桌面歌词偶尔会闪\n切歌的时候\n\n——\nCloudy ") {
		t.Errorf("body %q", body)
	}
	if link, _ := feedbackIssueURL("   "); link != feedbackIssues+"/choose" {
		t.Errorf("empty: %q", link)
	}
	link, long = feedbackIssueURL(strings.Repeat("很长的反馈", 600))
	if !long || len(link) > feedbackIssueMost || strings.Contains(link, "body=") {
		t.Errorf("long: %v, %d bytes", long, len(link))
	}
}

func TestFeedbackToIssue(t *testing.T) {
	oldCopy, oldOpen := copyText, openURL
	defer func() { copyText, openURL = oldCopy, oldOpen }()
	var copied, opened string
	copyText = func(s string) error { copied = s; return nil }
	openURL = func(u string) error { opened = u; return nil }
	a := newTestApp()
	a.feedback.text = "想要一个睡眠定时"
	a.feedbackToIssue()
	if !strings.HasPrefix(opened, feedbackIssues+"?") || copied != "" || a.feedback.failed {
		t.Errorf("opened %q copied %q note %q", opened, copied, a.feedback.note)
	}
	a.feedback.text = strings.Repeat("很长的反馈", 600)
	a.feedbackToIssue()
	if !strings.HasPrefix(copied, "很长的反馈") || a.feedback.failed {
		t.Errorf("long: copied %d bytes, note %q", len(copied), a.feedback.note)
	}
}

// The wall's file reads, empty names dropped; the one built in parses.
func TestParseSupport(t *testing.T) {
	d, err := parseSupport([]byte(`{"updates":37,"feedback":5,"supporters":[{"name":" 柚子 ","note":"加油"},{"name":"  "}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if d.Updates != 37 || d.Feedback != 5 || len(d.Supporters) != 1 || d.Supporters[0].Name != "柚子" {
		t.Errorf("%+v", d)
	}
	if _, err := supportFiles.ReadFile("support/supporters.json"); err != nil {
		t.Fatal(err)
	}
	if e := embeddedSupport(); e.Updates == 0 {
		t.Errorf("embedded %+v", e)
	}
	if codes := supportCodes(); codes[0] == nil || codes[1] == nil {
		t.Error("the pay codes built in do not decode")
	}
	if supportInitial("midnight") != "M" || supportInitial("") != "?" || supportInitial("云边") != "云" {
		t.Error("initials")
	}
}

// The card under 关于 opens the page in its place; the page shows the
// wall GitHub holds, and goes back; another category leaves it.
func TestSupportPage(t *testing.T) {
	a := newTestApp()
	// As read from GitHub: posts are dropped in tests.
	a.support = supportState{read: true, fetched: true, data: supportData{Updates: 38, Supporters: []supporter{{Name: "云边的猫", Note: "丝绸背景太好看了"}}}}
	a.settings = defaultSettings()
	a.settingsView.open, a.settingsView.tab = true, settingsAbout
	tt := ui.NewTester(a.view, 1180, 820)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	tt.Frame()
	if !tt.HasText("支持 Cloudy") || !tt.HasText("去支持") {
		t.Fatal("关于 has no support card")
	}
	if err := tt.Click("去支持"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(time.Second)
	tt.Frame()
	if !a.support.page || !tt.HasText("谢谢你愿意支持 Cloudy") || !tt.HasText("云边的猫") || !tt.HasText("38") {
		t.Fatal("the page does not show, or not the wall read")
	}
	if !tt.HasText("微信收款码\n还没有放") && !tt.HasText("扫码请 nya 喝杯奶茶") {
		t.Fatal("no codes")
	}
	_ = tt.Click("放大微信收款码")
	tt.Advance(time.Second)
	tt.Frame()
	if !a.support.large {
		t.Error("the code does not open large")
	}
	a.support.large = false
	tt.Advance(time.Second)
	tt.Frame()
	if err := tt.Click("返回关于"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.support.page {
		t.Error("back did not go back")
	}
	a.support.page = true
	tt.Frame()
	_ = tt.Click("播放")
	tt.Frame()
	if a.support.page {
		t.Error("another category kept the page")
	}
}

// The wall's fetch failing leaves the one built in.
func TestSupportWallKeepsEmbedded(t *testing.T) {
	old := fetchSupport
	defer func() { fetchSupport = old }()
	fetchSupport = func(context.Context) (supportData, error) { return supportData{}, errors.New("offline") }
	a := newTestApp()
	d := a.supportWall()
	time.Sleep(20 * time.Millisecond) // the fetch fails; without a window its post is dropped anyway
	if d.Updates != embeddedSupport().Updates {
		t.Errorf("%+v", *d)
	}
}

// 更新内容 offers the page, once, at its foot.
func TestWhatsNewOffersSupport(t *testing.T) {
	a := newTestApp()
	a.settings = defaultSettings()
	a.whatsNew = whatsNewState{open: true, version: "1.3.7", notes: "- 新增丝绸背景"}
	tt := ui.NewTester(a.view, 1180, 820)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	tt.Frame()
	if err := tt.Click("支持一下 nya"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.whatsNew.open || !a.support.page || !a.settingsView.open || a.settingsView.tab != settingsAbout {
		t.Error("the line does not open the support page")
	}
}
