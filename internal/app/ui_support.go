package app

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// supportEpoch is when the light of the page started drifting.
var supportEpoch = time.Now()

// The support page (支持 Cloudy): a card under 关于 opens it, in place of
// 关于's page, as does a line at the foot of 更新内容. It asks nothing of
// anyone elsewhere: no popup, no dot, nothing given in return.

var (
	supportIconStar = strokeIcon(`<path d="M12 3.2l2.7 5.6 6.1.8-4.5 4.2 1.1 6-5.4-2.9-5.4 2.9 1.1-6-4.5-4.2 6.1-.8z"/>`)
	supportIconTV   = strokeIcon(`<rect x="2.8" y="5.5" width="18.4" height="14" rx="3"/><path d="M7.5 2.8 10 5.5M16.5 2.8 14 5.5M9.5 11v3M14.5 11v3"/>`)
	supportIconNext = strokeIcon(`<path d="m9 6 6 6-6 6"/>`)
)

// supportMethods are the codes the page offers, and the tint of each.
var supportMethods = [2]struct {
	label, app string
	tint       ui.Color
}{
	{"微信", "微信", ui.Hex("#2fb36b")},
	{"支付宝", "支付宝", ui.Hex("#3a86f0")},
}

// supportAfdian is the author's page on 爱发电, where support can be
// monthly and come with a note.
const supportAfdian = "https://ifdian.net/a/nya12138"

// supportBilibili is the author's Bilibili space, where the videos are.
func supportBilibili() string {
	return fmt.Sprintf("https://space.bilibili.com/%d", feedbackBilibiliUID)
}

// supportGlow paints soft rounds of color drifting in r, clipped to its
// corners: the light of the card and of the page's head. t is in seconds.
func (a *app) supportGlow(p *ui.Painter, r ui.Rect, radius float32, strength float32, t float64) {
	accent := a.pal.accent
	warm, cool := ui.Hex("#e2965c"), ui.Hex("#965cb4")
	if !a.pal.dark {
		strength *= 0.7
	}
	blobs := []struct {
		x, y, size float32
		c          ui.Color
		a          float32
		period     float64
	}{
		{0.12, 0.05, 0.62, accent, 0.42, 11},
		{0.55, 0.85, 0.55, warm, 0.26, 14},
		{1.02, 0.0, 0.5, cool, 0.28, 17},
	}
	p.Clip(r, radius, func() {
		for i, b := range blobs {
			ph := t/b.period*2*math.Pi + float64(i)
			x := r.X + r.W*b.x + float32(math.Sin(ph))*18
			y := r.Y + r.H*b.y + float32(math.Cos(ph*0.8))*10
			size := max(r.W, r.H) * b.size * (1 + 0.06*float32(math.Sin(ph*1.3)))
			// A shadow of a dot is a round of light, soft all the way out.
			p.Shadow(ui.Rect{X: x, Y: y, W: 1, H: 1}, 0, 0, 0, size, size*0.12, b.c.Alpha(b.a*strength))
		}
	})
}

// supportMoves says whether the light drifts: in full motion, and not to
// save power.
func (a *app) supportMoves() bool {
	return a.motion == motionFull && !a.settings.PowerSaver
}

// supportCard is the card under 关于 that opens the page.
func (a *app) supportCard(k *settingsKit) {
	c, ink := k.c, k.ink
	wall := a.supportWall()
	row := ui.Row(c).Padding(18, 20).Gap(16).AlignItems(ui.Center)
	row.Draw(func(p *ui.Painter, r ui.Rect) {
		a.supportGlow(p, r, 14, 0.45, p.Now().Sub(supportEpoch).Seconds())
		if a.supportMoves() {
			p.AnimationFrameRate(20)
		}
	})
	row.Children(func() {
		heartTile(c, 52, 14, 24)
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(3).Children(func() {
			ui.Text(c, "支持 Cloudy").FontSize(14).FontWeight(700)
			ui.Text(c, "Cloudy 免费、开源、没有广告。如果它陪你听了很多歌，可以请 nya 喝杯奶茶，让它继续更新下去").
				FontSize(12).TextColor(ink.muted).LineHeight(1.5)
			// Who has, once someone has: nothing while the wall is empty.
			n := len(wall.Supporters)
			if n == 0 {
				return
			}
			ui.Row(c).Gap(8).AlignItems(ui.Center).Margin(7, 0, 0, 0).Children(func() {
				ui.Row(c).Children(func() {
					for i, s := range wall.Supporters[:min(n, 3)] {
						m := float32(-7)
						if i == 0 {
							m = 0
						}
						supportRound(c, s.Name, 22, 10).Margin(0, 0, 0, m)
					}
				})
				ui.Text(c, fmt.Sprintf("已有 %d 位朋友支持过", n)).FontSize(12).TextColor(ink.muted)
			})
		})
		if fillButton(c.Key("support-open"), "去支持", supportIconNext, a.pal.accent, a.pal.onAccent, false).Clicked() {
			a.support.page, a.support.all = true, false
			a.settingsView.scroll[settingsAbout].Y = 0
		}
	})
}

// heartTile is the heart on its round square, in the reds of the app.
func heartTile(c *ui.Context, size, radius, icon float32) {
	ui.Box(c).Size(size, size).Radius(radius).Center().Shrink(0).
		Shadow(0, 6, 18, 0, ui.Hex("#e5484d").Alpha(0.32)).
		Draw(func(p *ui.Painter, r ui.Rect) {
			p.FillGradient(r, ui.LinearGradient{From: ui.Hex("#f07a7e"), To: ui.Hex("#a52a40"), Angle: 145}, radius)
			p.Fill(ui.Rect{X: r.X + radius, Y: r.Y, W: r.W - 2*radius, H: 1}, white.Alpha(0.35), 0)
		}).
		Children(func() {
			ui.Icon(c, iconHeart).Size(icon, icon).TextColor(white)
		})
}

// supportRound is a name's first letter on a round of its color, ringed
// in the card's fill.
func supportRound(c *ui.Context, name string, size, font float32) ui.Element {
	return ui.Box(c).Size(size, size).Radius(999).Center().Shrink(0).Background(supportTint(name)).
		Border(2, ui.RGBA(0, 0, 0, 0.12)).Children(func() {
		ui.Text(c, supportInitial(name)).FontSize(font).FontWeight(700).TextColor(white)
	})
}

// supportPage is the page: why, what came of it, the codes, the wall, and
// what helps besides money.
func (a *app) supportPage(k *settingsKit) {
	c, ink := k.c, k.ink
	st := &a.support
	wall := a.supportWall()
	fill, edge := a.cardColors()

	back := ui.ButtonBase(c.Key("support-back")).Height(28).PaddingX(10).Gap(4).Radius(999).AlignItems(ui.Center).
		AlignSelf(ui.Start).FocusRing(false).Label("返回关于").Margin(0, 0, 6, -6)
	bg := ui.Color{}
	if back.Hovered() {
		bg = ink.hover
	}
	back.Background(back.AnimateColor("bg", bg, motionHover)).TextColor(ink.muted).Children(func() {
		ui.Icon(c, iconBack).Size(14, 14)
		ui.Text(c, "关于").FontSize(13)
	})
	if back.Clicked() {
		st.page = false
	}
	ui.Text(c, "支持 Cloudy").FontSize(26).FontWeight(700).Margin(0, 40, 0, 4)
	ui.Text(c, "与网易云音乐无关，赞助不换取任何功能。").FontSize(13).TextColor(ink.muted).Margin(6, 40, 0, 4)

	// The head: the light drifting behind the heart, and a word from nya.
	head := ui.Row(c.Key("support-head")).Margin(24, 0, 0, 0).Padding(26, 28).Gap(22).AlignItems(ui.Center).
		Radius(20).Background(fill).Border(0.5, edge).MinHeight(170)
	head.Draw(func(p *ui.Painter, r ui.Rect) {
		a.supportGlow(p, r, 20, 1, p.Now().Sub(supportEpoch).Seconds())
		if a.supportMoves() {
			p.AnimationFrameRate(30)
		}
	})
	head.Children(func() {
		heartTile(c, 84, 26, 40)
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(8).Children(func() {
			ui.Text(c, "谢谢你愿意支持 Cloudy").FontSize(22).FontWeight(700)
			ui.Text(c, "Cloudy 是我业余时间一点点做出来的，一直免费、没有广告。你的支持会变成更多的更新、更快的修复，也让我知道有人在用它听歌。").
				FontSize(13).LineHeight(1.75).TextColor(ink.text.Alpha(0.86)).MaxWidth(520)
			ui.Text(c, "— nya").FontSize(12).TextColor(ink.muted)
		})
	})

	// What came of it, as the wall's file counts it: 0 hides a number.
	type stat struct{ n, unit, desc string }
	var stats []stat
	if wall.Updates > 0 {
		stats = append(stats, stat{fmt.Sprint(wall.Updates), "次更新", "从 1.1 开始，发布在 GitHub"})
	}
	if wall.Feedback > 0 {
		stats = append(stats, stat{fmt.Sprint(wall.Feedback), "条反馈", "来自 B 站评论、私信和 GitHub"})
	}
	stats = append(stats, stat{"0", "广告 · 付费功能", "赞助不会换来任何功能，以后也不会"})
	ui.Row(c).Gap(12).Margin(12, 0, 0, 0).AlignItems(ui.Stretch).Children(func() {
		for _, s := range stats {
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(3).Padding(12, 16).Radius(14).Background(fill).Border(0.5, edge).Children(func() {
				ui.Row(c).Gap(4).AlignItems(ui.Baseline).Children(func() {
					ui.Text(c, s.n).FontSize(20).FontWeight(700).FontFeatures("tnum")
					ui.Text(c, s.unit).FontSize(13).FontWeight(500).TextColor(ink.text.Alpha(0.8))
				})
				ui.Text(c, s.desc).FontSize(12).TextColor(ink.muted).LineHeight(1.4)
			})
		}
	})

	// The codes.
	m := supportMethods[st.method]
	codes := supportCodes()
	ui.Row(c).Margin(12, 0, 0, 0).Padding(20, 22).Gap(26).AlignItems(ui.Center).Radius(20).Background(fill).Border(0.5, edge).Children(func() {
		ui.Column(c).Gap(12).AlignItems(ui.Center).Shrink(0).Children(func() {
			a.segmentedControl(c.Key("support-method"), &st.method, []string{supportMethods[0].label, supportMethods[1].label}, true, nil)
			code := ui.ButtonBase(c.Key("support-code")).Padding(8).Radius(14).FocusRing(false).Label("放大" + m.label + "收款码").Cursor(ui.CursorPointer)
			code.Background(code.AnimateColor("tint", m.tint, ui.Animation{Duration: 300e6})).
				Shadow(0, 8, 24, 0, m.tint.Alpha(0.25)).Children(func() {
				supportCodeFace(c, codes[st.method], m.label, 144)
			})
			if code.Clicked() {
				a.pressedFrom(code, 14)
				st.large = true
			}
		})
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(14).Children(func() {
			ui.Text(c, "扫码请 nya 喝杯奶茶").FontSize(16).FontWeight(700)
			ui.Column(c).Gap(10).Children(func() {
				for i, line := range []string{
					"打开" + m.app + "，扫左边的码",
					"金额随意，一块钱也是心意",
					"想上感谢墙？付款时备注一个昵称，或者在 B 站私信 nya",
				} {
					ui.Row(c).Gap(10).AlignItems(ui.Start).Children(func() {
						ui.Box(c).Size(20, 20).Radius(999).Center().Shrink(0).Margin(1, 0, 0, 0).Background(a.pal.accent.Alpha(0.16)).Children(func() {
							ui.Text(c, fmt.Sprint(i+1)).FontSize(11).FontWeight(700).TextColor(a.pal.accent)
						})
						ui.Text(c, line).FontSize(13).LineHeight(1.55).TextColor(ink.text.Alpha(0.88)).Shrink(1).MinWidth(0)
					})
				}
			})
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				if k.button("support-large", "放大收款码", iconExpand, false, false).Clicked() {
					st.large = true
				}
				if k.button("support-afdian", "去爱发电", iconExternal, false, false).
					Tooltip("在爱发电支持：可以按月赞助，留言写个昵称就能上感谢墙").Clicked() {
					openLink(supportAfdian)
				}
			})
		})
	})

	// The wall.
	ui.Row(c).Margin(24, 16, 8, 16).AlignItems(ui.Baseline).Children(func() {
		ui.Text(c, "感谢墙").FontSize(12.5).FontWeight(600).TextColor(ink.muted).Grow(1)
		ui.Text(c, "只列同意公开的朋友").FontSize(12).TextColor(ink.muted.Alpha(0.8))
	})
	ui.Column(c).Padding(18, 20, 16, 20).Gap(14).Radius(20).Background(fill).Border(0.5, edge).Children(func() {
		n := len(wall.Supporters)
		ui.Row(c).AlignItems(ui.Baseline).Children(func() {
			ui.Text(c, "这些朋友让 Cloudy 走到了今天").FontSize(15).FontWeight(700).Grow(1)
			if n > 0 {
				ui.Text(c, fmt.Sprintf("共 %d 位 · 最新的在前", n)).FontSize(12).TextColor(ink.muted)
			}
		})
		if n == 0 {
			ui.Text(c, "还没有名字。第一个支持的朋友，会出现在这里。").FontSize(13).TextColor(ink.muted).Margin(4, 0, 6, 0)
			return
		}
		shown := wall.Supporters
		if !st.all && len(shown) > supportWallRows {
			shown = shown[:supportWallRows]
		}
		ui.Grid(c).Columns(3).GapX(10).GapY(10).Children(func() {
			for i, s := range shown {
				ui.Row(c.Key(i)).Padding(12, 14).Gap(10).Radius(14).Background(ink.hover).MinWidth(0).Children(func() {
					supportRound(c, s.Name, 30, 12)
					ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(2).Children(func() {
						ui.Text(c, s.Name).FontSize(13).FontWeight(500).SingleLine()
						note, col := s.Note, ink.text.Alpha(0.72)
						if note == "" {
							note, col = "没有留言", ink.muted
						}
						ui.Text(c, note).FontSize(12).LineHeight(1.5).TextColor(col).MaxLines(2)
					})
				})
			}
		})
		if len(wall.Supporters) > supportWallRows {
			label := fmt.Sprintf("查看全部 %d 位", len(wall.Supporters))
			if st.all {
				label = "收起"
			}
			ui.Row(c).Justify(ui.Center).Children(func() {
				if k.button("support-all", label, nil, false, false).Clicked() {
					st.all = !st.all
				}
			})
		}
	})

	// What helps besides money.
	k.group("不付钱也能帮忙", func() {
		k.stateRow("给仓库点个 Star", "在 GitHub 的 "+releaseRepo+"，让更多人看到它", func() {
			if k.button("support-star", "打开 GitHub", supportIconStar, false, false).Clicked() {
				openLink(appSource)
			}
		})
		k.stateRow("在 B 站一键三连", "新版本的介绍视频都发在那里", func() {
			if k.button("support-bilibili", "打开 B 站", supportIconTV, false, false).Clicked() {
				openLink(supportBilibili())
			}
		})
		k.stateRow("告诉我哪里不好用", "很多新功能都来自大家的反馈", func() {
			if k.button("support-feedback", "去反馈", iconComment, false, false).Clicked() {
				st.page = false
				a.settingsView.scroll[settingsAbout].Y = math.MaxFloat32
			}
		})
	})
}

// supportCodeFace is a pay code on white, or where the app has none, a
// box that says so.
func supportCodeFace(c *ui.Context, bmp *ui.Bitmap, label string, size float32) {
	ui.Box(c).Size(size, size).Radius(8).Background(white).Center().Children(func() {
		if bmp != nil {
			// A wider white edge than the code's own, for phones to find it.
			ui.Image(c, bmp).Size(size*0.88, size*0.88).Fit(ui.Contain)
			return
		}
		dark := ui.Hex("#1d1d1f")
		ui.Box(c).Fill().Draw(func(p *ui.Painter, r ui.Rect) {
			f := size * 0.22
			for _, at := range [][2]float32{{0, 0}, {1, 0}, {0, 1}} {
				x := r.X + 8 + at[0]*(r.W-16-f)
				y := r.Y + 8 + at[1]*(r.H-16-f)
				p.Fill(ui.Rect{X: x, Y: y, W: f, H: f}, dark, 6)
				p.Fill(ui.Rect{X: x + 5, Y: y + 5, W: f - 10, H: f - 10}, white, 3)
				p.Fill(ui.Rect{X: x + 9, Y: y + 9, W: f - 18, H: f - 18}, dark, 2)
			}
		})
		ui.Text(c, label+"收款码\n还没有放").FontSize(12).LineHeight(1.5).TextColor(ui.Hex("#55555a")).TextAlign(ui.Center).Absolute()
	})
}

// supportCodeDialog shows the code chosen large, for a phone held to it.
func (a *app) supportCodeDialog(c *ui.Context) {
	st := &a.support
	if !a.settingsView.open {
		st.large = false
	}
	a.dialog(c, "support-code-large", 0.35, &st.large, func(panel ui.Element) {
		ink := a.inkOn(false)
		m := supportMethods[st.method]
		panel.Width(400).TextColor(ink.text).Label(m.label + "收款码")
		a.dialogPane(panel, 26, false)
		panel.Children(func() {
			dialogHeader(c, ink, "用"+m.app+"扫一扫", "金额随意，谢谢你", func() { st.large = false })
			ui.Row(c).Justify(ui.Center).Padding(4, 28, 0, 28).Children(func() {
				ui.Box(c).Padding(10).Radius(20).Background(m.tint).Children(func() {
					supportCodeFace(c, supportCodes()[st.method], m.label, 300)
				})
			})
			dialogFooter(c, func() {
				other := supportMethods[1-st.method]
				if settingButton(c.Key("switch"), ink, "换成"+other.label, nil, false, false).Clicked() {
					st.method = 1 - st.method
				}
				if primaryButton(c.Key("ok"), "好的", nil, a.pal, false).Clicked() {
					st.large = false
				}
			})
		})
	})
}

// supportLine is the line at the foot of 更新内容 that opens the page.
func (a *app) supportLine(c *ui.Context, ink paneInk, then func()) {
	b := ui.ButtonBase(c.Key("support-line")).Height(32).Gap(6).AlignItems(ui.Center).Grow(1).Shrink(1).MinWidth(0).
		FocusRing(false).Label("支持一下 nya").Cursor(ui.CursorPointer)
	col := ink.muted
	if b.Hovered() {
		col = a.pal.accent
	}
	b.TextColor(b.AnimateColor("ink", col, motionHover)).Children(func() {
		ui.Icon(c, iconHeart).Size(13, 13).TextColor(a.pal.accent)
		ui.Text(c, "喜欢这次更新？支持一下 nya").FontSize(12.5).SingleLine().Shrink(1).MinWidth(0)
		ui.Icon(c, supportIconNext).Size(11, 11)
	})
	if b.Clicked() {
		then()
	}
}

// supportWords are what a search of the settings finds the card by.
var supportWords = strings.Join([]string{"支持", "赞助", "捐赠", "打赏", "感谢墙", "收款码", "奶茶", "Cloudy"}, " ")
