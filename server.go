package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"net"
	"net/http"
	"net/http/pprof"
	"runtime"
	// "time"
)

const (
	DB_USER   = "loguser"
	DB_PASSWD = "loguser"
	DB_URL    = "192.168.8.78:3306"
)

type json_struct struct {
	RQH      string
	RPH      string
	UA       string
	BODY     string
	NET      string
	SP       string
	AREA     string
	OPENUID  string
	USERID   string
	ECODE    string
	EMSG     string
	IP       string
	PORT     string
	CLIENTIP string
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	m := martini.Classic()
	// m.Use(gzip.All())
	//=====================================================================
	m.Get("/debug/pprof", pprof.Index)
	m.Get("/debug/pprof/cmdline", pprof.Cmdline)
	m.Get("/debug/pprof/profile", pprof.Profile)
	m.Get("/debug/pprof/symbol", pprof.Symbol)
	m.Post("/debug/pprof/symbol", pprof.Symbol)
	m.Get("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
	m.Get("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	m.Get("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	m.Get("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	// m.Get("/stats", getStats)
	//=====================================================================
	m.Post("/clog", getPost)
	m.Get("/mlog", getParams)
	http.ListenAndServe(":80", m)
	m.Run()
}

func getParams(req *http.Request) (int, string) {
	// var response string
	var t json_struct
	var ipstr string
	// response = "GET\n\n"
	v := req.URL.Query()
	t.RQH = v.Get("RQH")
	t.RPH = v.Get("RPH")
	t.UA = v.Get("UA")
	t.BODY = v.Get("BODY")
	t.NET = v.Get("NET")
	t.SP = v.Get("SP")
	t.AREA = v.Get("AREA")
	t.OPENUID = v.Get("OPENUID")
	t.USERID = v.Get("USERID")
	t.ECODE = v.Get("ECODE")
	t.EMSG = v.Get("EMSG")
	ip, port, err2 := net.SplitHostPort(req.RemoteAddr)
	if err2 != nil {
		return 501, err2.Error()
	}
	t.IP = ip
	t.PORT = port
	iplist := req.Header["X-Forwarded-For"]
	for i, ip := range iplist {
		if i > 0 {
			ipstr += "#"
		}
		ipstr += ip
	}
	t.CLIENTIP = ipstr

	go t.insertDB()
	flog("============get=return=======================")
	return 200, "{\"returncode\":0,\"message\":\"success\"}"
}
func getPost(w http.ResponseWriter, r *http.Request) (int, string) {
	var t json_struct
	var ipstr string
	for key, values := range r.Header {
		flogn(key + "=====")
		for _, value := range values {
			flogn(value + ":")
		}
	}
	errorjson := r.FormValue("errorjson")
	flog("==========  start ===============")
	flog(errorjson)

	flog("==========  end  ===============")
	if errorjson == "" {
		return 509, "{\"returncode\":-1,\"message\":\"empty\"}"
	}
	//decoder := json.NewDecoder(errorjson)
	//decoder := json.NewDecoder(r.Body)
	// err1 := decoder.Decode(&t)
	err1 := json.Unmarshal([]byte(errorjson), &t)
	if err1 != nil {
		// panic()
		return 501, "{\"returncode\":-1,\"message\":\"" + err1.Error() + "\"}"
	}
	iplist := r.Header["X-Forwarded-For"]
	for i, ip := range iplist {
		if i > 0 {
			ipstr += "#"
		}
		ipstr += ip
	}
	t.CLIENTIP = ipstr
	// log.Println(t.Test)
	//t.IP = r.RemoteAddr
	ip, port, err2 := net.SplitHostPort(r.RemoteAddr)
	if err2 != nil {
		return 501, "{\"returncode\":-1,\"message\":\"" + err2.Error() + "\"}"
	}
	t.IP = ip
	t.PORT = port
	go t.insertDB()
	flog(t)
	flog("============post=return=======================")
	return 200, "{\"returncode\":0,\"message\":\"success\"}"
}
func (r *json_struct) insertDB() error {
	flog("=========start insert===============")
	var err error
	defer handlerError(&err)
	db, e := connFactory()
	defer db.Close()
	if e != nil {
		flog("==============1ERROR?")
		flog(e)
		return e
	}
	st, err := db.Prepare(`INSERT INTO dnslog.loginfo (RQH,RPH,UA,BODY,NET,SP,AREA,OPENUID,USERID,ECODE,EMSG,IP,PORT,CLIENTIP) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		flog("==============2ERROR?")
		flog(err)
		return err
	}
	_, err = st.Exec(r.RQH, r.RPH, r.UA, r.BODY, r.NET, r.SP, r.AREA, r.OPENUID, r.USERID, r.ECODE, r.EMSG, r.IP, r.PORT, r.CLIENTIP)
	if err != nil {
		flog("==============3ERROR?")
		flog(err)
		//flog(a)
		return err
	}
	flog("===========insert is  Done================")
	return nil
}
func connFactory() (*sql.DB, error) {
	db, e := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/dnslog?charset=utf8", DB_USER, DB_PASSWD, DB_URL))
	return db, e
}
func handlerError(err *error) {
	if err := recover(); err != nil {
		println(err)
	}
}
func flog(str interface{}) {
	fmt.Println(str)
}
func flogn(str interface{}) {
	fmt.Print(str)
}
