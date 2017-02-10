package main

import (

	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"html/template"
	_ "github.com/go-sql-driver/mysql"
	sql "database/sql"
	"github.com/siddontang/go/log"
	"strconv"
//	"regexp"
//	"encoding/json"
//	"time"
//	"github.com/garyburd/redigo/redis"
	"github.com/Unknwon/goconfig"
//	"strconv"
	"bytes"
//	"os"
//	"bufio"
//	"io"
//	"strings"
	m "model"
	u "utils"
)

type FileData struct {
	ID int
	UID int
	Title string
}

var db *sql.DB
var err error
var username, password, url, address, redis_Pwd, mode, logLevel, redis_db string
var redis_Database int
var ConfError error
var cfg *goconfig.ConfigFile

//Mysql Redis初始化
func Init() {
	cfg, ConfError = goconfig.LoadConfigFile("config.ini")
	if ConfError != nil {
		log.Error("配置文件config.ini不存在,请将配置文件复制到运行目录下")
	}
	logLevel, ConfError = cfg.GetValue("Log", "logLevel")
	if ConfError != nil {
		log.SetLevel(log.LevelInfo)
	} else {
		log.SetLevelByName(logLevel)
	}
	username, ConfError = cfg.GetValue("MySQL", "username")
	if ConfError != nil {
		log.Error("读取数据库username错误")
	}
	password, ConfError = cfg.GetValue("MySQL", "password")
	if ConfError != nil {
		log.Error("读取数据库password错误")
	}
	url, ConfError = cfg.GetValue("MySQL", "url")
	if ConfError != nil {
		log.Error("读取数据库url错误")
	}

	var dataSourceName bytes.Buffer
	dataSourceName.WriteString(username)
	dataSourceName.WriteString(":")
	dataSourceName.WriteString(password)
	dataSourceName.WriteString("@")
	dataSourceName.WriteString(url)
	db, err = sql.Open("mysql", dataSourceName.String())
	if err != nil {
		log.Error(err.Error())
	}
	if err := db.Ping(); err != nil {
		panic("Error Connection database...")
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(30)

	u.LISTMAX = 300
	u.PAGEMAX = 20
	u.NAVMAX = 5
	u.InitCateMap()
}



func Index(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("Hello, World!"))
//	fds := GetFileData()
	lp := m.GenerateListPageVar(db, 0, 1)
	render(w, "templates/index.html", lp)
}

func Greet(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	w.Write([]byte(fmt.Sprintf("Hello %s !", name)))
}

func ListShare(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	cat := vars["category"]
	cati, ok:= u.CAT_STR_INT[cat]
	if !ok {
		log.Info(err)
		return
	}

	p := vars["page"]
	if p == "" {
		p = "1"
	}
	pp, err:=strconv.Atoi(p)
	if err != nil {
		log.Info(err )
		return
	}
	pv := m.GenerateListPageVar(db, cati, pp)
	if pv != nil {
		render(w, "templates/index.html", pv)
	}
}

func ListUsers(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	p := vars["page"]
	if p == "" {
		p = "1"
	}
	pp, err:=strconv.Atoi(p)
	if err != nil {
		log.Info(err )
		return
	}
	pv := m.GenerateUlistPageVar(db, pp)
	if pv != nil {
		render(w, "templates/index.html", pv)
	}
}


func SearchShare(w http.ResponseWriter, r *http.Request) {

	// break down the variables for easier assignment
	vars := mux.Vars(r)
	t := vars["category"]
	kw := vars["keyword"]
	w.Write([]byte(fmt.Sprintf("Category is %s ", t)))
	w.Write([]byte(fmt.Sprintf("keyword is %s ", kw)))
}

func ShowShare(w http.ResponseWriter, r *http.Request) {

	// break down the variables for easier assignment
	vars := mux.Vars(r)
	id := vars["dataid"]
	sp := m.GenerateShareVar(db, id)
	if sp != nil {
		render(w, "templates/index.html", sp)
	}
}

func ShowUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uk := vars["uk"]
	p := vars["page"]
	if p == "" {
		p = "1"
	}

	pp, err:=strconv.Atoi(p)
	if err != nil {
		log.Info(err )
		return
	}

	pv := m.GenerateUserPageVar(db, uk, pp)
	if pv != nil {
		render(w, "templates/index.html", pv)
	}
}


func render(w http.ResponseWriter, filename string, data interface{}) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Error(err)
}

func GetFileData() ([]FileData) {
	db, err := sql.Open("mysql", "root@/baidu?charset=utf8")
	u.CheckErr(err)

	// query
	rows, err := db.Query("SELECT id, uinfo_id, title FROM sharedata limit 0, 100")
	u.CheckErr(err)

	fds := []FileData{}

	for rows.Next() {
		var id int
		var uid int
		var title string
		err = rows.Scan(&id, &uid, &title)
		u.CheckErr(err)
		fd := FileData {
			ID : id,
			UID : uid,
			Title : title,
		}
		fds = append(fds, fd)
	}
	return fds
}

func main() {

	Init()

	mx := mux.NewRouter()

	mx.HandleFunc("/", Index)
	//list
	mx.HandleFunc("/list/{category}", ListShare)
	mx.HandleFunc("/list/{category}/", ListShare)
	mx.HandleFunc("/list/{category}/{page}", ListShare)
	mx.HandleFunc("/list/{category}/{page}/", ListShare)

	//ulist
	mx.HandleFunc("/ulist}", ListUsers)
	mx.HandleFunc("/ulist}/", ListUsers)
	mx.HandleFunc("/ulist}/{page}", ListUsers)
	mx.HandleFunc("/ulist}/{page}/", ListUsers)

	//search
	mx.HandleFunc("/search/{category}/{keyword}", SearchShare)
	mx.HandleFunc("/search/{category}/{keyword}/", SearchShare)

	//file
	mx.HandleFunc("/file/{dataid}", ShowShare)
	mx.HandleFunc("/file/{dataid}/", ShowShare)

	//user
	mx.HandleFunc("/user/{uk}", ShowUser)
	mx.HandleFunc("/user/{uk}/", ShowUser)
	mx.HandleFunc("/user/{uk}/{page}", ShowUser)
	mx.HandleFunc("/user/{uk}/{page}/", ShowUser)
	//server static
	mx.PathPrefix("/static").Handler(http.FileServer(http.Dir("./")))

	log.Info("Listening...")
	http.ListenAndServe(":8080", mx)

}
