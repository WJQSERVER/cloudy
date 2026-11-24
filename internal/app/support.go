package app

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo/ui"
)

// What the support page holds that is not code: the thanks wall and the
// pay codes, in support/ (its README says how to fill it). The wall is
// read from the repository too, so that a name added there shows without
// a new version; the codes are only what the app was built with.
//
//go:embed support
var supportFiles embed.FS

// supportData is support/supporters.json: the wall, and the two numbers
// the page shows of the work (0 hides one).
type supportData struct {
	Updates    int         `json:"updates"`
	Feedback   int         `json:"feedback"`
	Supporters []supporter `json:"supporters"`
}

// A supporter is a name on the wall, newest first, and what they said.
type supporter struct {
	Name string `json:"name"`
	Note string `json:"note,omitempty"`
}

// supportRemote is where the wall is kept, read straight from GitHub or
// through the update mirrors, which serve its raw files as well.
const supportRemote = "https://raw.githubusercontent.com/" + releaseRepo + "/main/internal/app/support/supporters.json"

// supportState is the support page: shown in place of 关于, the code
// chosen, the wall as read, and the code shown large.
type supportState struct {
	page    bool
	method  int // 0 微信, 1 支付宝
	data    supportData
	read    bool // data holds the embedded wall, at least
	fetched bool // the wall was asked of GitHub, this run
	all     bool // the whole wall shows, not its first rows
	large   bool // the code shows large, in a dialog
}

// supportWallRows is how many names the wall shows before 查看全部.
const supportWallRows = 9

// parseSupport reads supporters.json, dropping names left empty.
func parseSupport(b []byte) (supportData, error) {
	var d supportData
	if err := json.Unmarshal(b, &d); err != nil {
		return supportData{}, err
	}
	kept := d.Supporters[:0]
	for _, s := range d.Supporters {
		s.Name, s.Note = strings.TrimSpace(s.Name), strings.TrimSpace(s.Note)
		if s.Name != "" {
			kept = append(kept, s)
		}
	}
	d.Supporters = kept
	return d, nil
}

// embeddedSupport is the wall the app was built with.
func embeddedSupport() supportData {
	b, err := supportFiles.ReadFile("support/supporters.json")
	if err != nil {
		return supportData{}
	}
	d, err := parseSupport(b)
	if err != nil {
		log.Printf("support: %v", err)
	}
	return d
}

// fetchSupport asks for the wall as it is now, from GitHub and then each
// mirror; tests swap it.
var fetchSupport = func(ctx context.Context) (supportData, error) {
	var last error
	for _, prefix := range append([]string{""}, updateMirrors...) {
		d, err := fetchSupportFrom(ctx, prefix+supportRemote)
		if err == nil {
			return d, nil
		}
		last = err
		if ctx.Err() != nil {
			break
		}
	}
	return supportData{}, last
}

func fetchSupportFrom(ctx context.Context, url string) (supportData, error) {
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return supportData{}, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return supportData{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return supportData{}, fmt.Errorf("support: %s", res.Status)
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return supportData{}, err
	}
	return parseSupport(b)
}

// supportWall is the wall to show: the embedded one at first, then, once,
// what GitHub holds, which replaces it when it is read.
func (a *app) supportWall() *supportData {
	st := &a.support
	if !st.read {
		st.read, st.data = true, embeddedSupport()
	}
	// Without a window, as in tests, nothing would take what is read.
	if !st.fetched && a.win != nil {
		st.fetched = true
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			d, err := fetchSupport(ctx)
			if err != nil {
				log.Printf("support: %v", err)
				return
			}
			a.post(func() {
				a.support.data = d
				a.redraw()
			})
		}()
	}
	return &st.data
}

// supportCodes are the pay codes, nil where the app has none.
var supportCodes = sync.OnceValue(func() [2]*ui.Bitmap {
	var out [2]*ui.Bitmap
	for i, name := range []string{"wechat.png", "alipay.png"} {
		b, err := supportFiles.ReadFile("support/" + name)
		if err != nil {
			continue
		}
		img, _, err := image.Decode(bytes.NewReader(b))
		if err != nil {
			log.Printf("support: %s: %v", name, err)
			continue
		}
		out[i] = ui.NewBitmap(img)
	}
	return out
})

// openSupport shows the support page, in the settings under 关于.
func (a *app) openSupport() {
	a.support.page, a.support.all = true, false
	a.openSettingsAt(settingsAbout)
}

// supportInitial is the letter of a name its round shows.
func supportInitial(name string) string {
	for _, r := range name {
		if r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		return string(r)
	}
	return "?"
}

// supportTints are the colors of the rounds, chosen by the name.
var supportTints = []ui.Color{ui.Hex("#5b7bd5"), ui.Hex("#c58a3a"), ui.Hex("#4a9a7a"), ui.Hex("#9a5bb5"), ui.Hex("#d0605e"), ui.Hex("#3f8fb0")}

func supportTint(name string) ui.Color {
	h := uint32(2166136261)
	for _, b := range []byte(name) {
		h = (h ^ uint32(b)) * 16777619
	}
	return supportTints[h%uint32(len(supportTints))]
}
