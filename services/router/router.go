package router

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	re_error           = regexp.MustCompile(`var sPageErr = (msgLoginDup_err|msgLoginPwd_err);`)
	re_notLoggedIn     = regexp.MustCompile(`bShowProtection=1`)
	re_diagOnline      = regexp.MustCompile(`var online_time='([\d\:]+)'`)
	re_diagSpeed       = regexp.MustCompile(`>(\d+)&nbsp;\(Kbps\.\)`)
	re_diagNoiseMargin = regexp.MustCompile(`>([\d\.]+)&nbsp;dB`)

	ErrLoginFailed       = errors.New("Router login failed, check password/duplicate admin")
	ErrNotLoggedIn       = errors.New("Router not logged in")
	ErrDiagNoOnline      = errors.New("No online time found in router diagnostic")
	ErrDiagNoSpeed       = errors.New("No speed found in router diagnostic")
	ErrDiagNoNoiseMargin = errors.New("No noise margin found in router diagnostic")
)

type Config struct {
	IP       string
	Password string
}

type Router struct {
	cfg      Config
	lock     sync.Mutex
	http     http.Client
	loggedIn bool
}

type Diagnostic struct {
	Online          string
	UpSpeed         int
	UpNoiseMargin   float32
	DownSpeed       int
	DownNoiseMargin float32
	Timestamp       time.Time
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("%d kbps | %.1f dB | %.0f%%", d.UpSpeed, d.UpNoiseMargin, d.Quality()*100)
}

func (d Diagnostic) Quality() float32 {
	q := (float32(d.DownSpeed) + 1200/3*(float32(d.DownNoiseMargin)-6)) / 12000
	if q < 0 {
		return 0
	}
	return q
}

func New(cfg Config) *Router {
	return &Router{
		cfg: cfg,
	}
}

func (r *Router) Login() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.loggedIn = false

	v := url.Values{}
	v.Set("controller", "Overview")
	v.Set("Action", "Login")
	v.Set("id", "0")
	v.Set("idTextPassword", r.cfg.Password)
	resp, err := r.http.PostForm("http://"+r.cfg.IP+"/cgi-bin/Hn_login.cgi", v)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if re_error.Match(body) {
		return ErrLoginFailed
	}

	r.loggedIn = true
	return nil
}

func (r *Router) Logout() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	_, err := r.http.Get("http://" + r.cfg.IP + "/cgi-bin/Hn_logout.cgi?controller=Overview&action=Logout")
	if err != nil {
		return err
	}
	r.loggedIn = false
	return nil
}

func (r *Router) IsLoggedIn() bool {
	return r.loggedIn
}

func (r *Router) Diagnostic() (Diagnostic, error) {
	diag := Diagnostic{}
	if !r.IsLoggedIn() {
		return diag, ErrNotLoggedIn
	}

	resp, err := r.http.Get("http://" + r.cfg.IP + "/diagnostic.htm")
	if err != nil {
		return diag, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return diag, err
	}

	m := re_diagOnline.FindSubmatch(body)
	if m == nil {
		return diag, ErrDiagNoOnline
	}
	diag.Online = string(m[1])

	mm := re_diagSpeed.FindAllSubmatch(body, -1)
	if mm == nil {
		return diag, ErrDiagNoSpeed
	}
	diag.UpSpeed, _ = strconv.Atoi(string(mm[0][1]))
	diag.DownSpeed, _ = strconv.Atoi(string(mm[1][1]))

	mm = re_diagNoiseMargin.FindAllSubmatch(body, -1)
	if mm == nil {
		return diag, ErrDiagNoNoiseMargin
	}
	f, _ := strconv.ParseFloat(string(mm[0][1]), 32)
	diag.UpNoiseMargin = float32(f)
	f, _ = strconv.ParseFloat(string(mm[1][1]), 32)
	diag.DownNoiseMargin = float32(f)

	diag.Timestamp = time.Now()
	return diag, nil
}
