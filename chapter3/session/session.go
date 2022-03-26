package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var globalSession *session.Manager

func init() {
	globalSession, _ = session.NewManager("memory", "gosessionid", 3600)
	go globalSession.GC()
}

type Session interface {
	Set(key, value interface{}) error //设置session
	Get(key interface{}) interface{}  //获取session
	Delete(key interface{}) error     //删除session
	SessionID() string                //返回sessionId
}

type Provider interface {
	SessionInit(sessionId string) (Session, error)
	SessionRead(sessionId string) (Session, error)
	SessionDestory(sessionId string) error
	GarbageCollector(maxLifeTime int64)
}

var providers = make(map[string]Provider)

func RegisterProvider(name string, provider Provider) {
	if provider == nil {
		panic("session: Register provider is nil")
	}

	if _, p := providers[name]; p {
		panic("session: Register provider is existed")
	}
	providers[name] = provider
}

type SessionManager struct {
	cookieName  string     //cookie的名称
	lock        sync.Mutex //锁，保证并发时数据的安全性和一致性
	provider    Provider   //管理session
	maxLifeTime int64      //超时时间
}

func NewSessionManager(providerName, cookieName string, maxLifetime int64) (*SessionManager, error) {
	provider, ok := providers[providerName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", providerName)
	}

	return &SessionManager{
		cookieName:  cookieName,
		maxLifeTime: maxLifetime,
		provider:    provider,
	}, nil
}

func (manager *SessionManager) GetSessinonId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (manager *SessionManager) SessionBegin(w http.ResponseWriter, r *http.Request) (session Session) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		sessionId := manager.GetSessinonId()
		session, _ = manager.provider.SessionInit(sessionId)
		cookie := http.Cookie{
			Name:     manager.cookieName,
			Value:    url.QueryEscape(sessionId),
			Path:     "/",
			HttpOnly: true,
			MaxAge:   int(manager.maxLifeTime),
		}
		http.SetCookie(w, &cookie)
	} else {
		sessionId, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.provider.SessionRead(sessionId)
	}
	return session
}

func (manager *SessionManager) SessionDestory(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	manager.provider.SessionDestory(cookie, Value)
	expiredTime := time.Now()
	newCookie := http.Cookie{
		Name:     manager.cookieName,
		Path:     "/",
		HttpOnly: true,
		Expires:  expiredTime,
		MaxAge:   -1,
	}
	http.SetCookie(w, &newCookie)
}

/*
func init(){
	go globalSession.GarbageCollector()
}

*/

func (manager *SessionManager) GarbageCollector() {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.provider.GarbageCollector(manager.maxLifeTime)
	time.AfterFunc(time.Duration(manager.maxLifeTime), func() {
		manager.GarbageCollector()
	})
}

func login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "GET" {
		t, _ := template.ParseFiles("chapter3/session/login.html")
		w.Header().Set("Content-Type", "text/html")
		t.Execute(w, "aaaa")
	} else {
		sess := globalSession.SessionStart(w, r)
		sess.Set("phone", r.Form["phone"])
	}
}

func main() {
	http.HandleFunc("/login", login)
	err := http.ListenAndServe(":8088", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
