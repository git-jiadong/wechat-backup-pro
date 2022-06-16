package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/greycodee/wcdb-parse/db"
)

var enmicromsg *db.EnMicroMsg
var wxfileindex *db.WxFileIndex

//go:embed static
var htmlFile embed.FS

var serverPort = "8080"

func main() {
	enmicromsg = db.OpenEnMicroMsg("/mnt/d/MicroMsg/enmicromsg_plaintext.db")
	wxfileindex = db.OpenWxFileIndex("/mnt/d/MicroMsg/wxfileindex_plaintext.db")

	fsys, _ := fs.Sub(htmlFile, "static")
	staticHandle := http.FileServer(http.FS(fsys))

	// 文件路由
	fs := http.FileServer(http.Dir("/mnt/d/MicroMsg/"))
	http.Handle("/media/", http.StripPrefix("/media/", fs))

	http.Handle("/", staticHandle)
	http.Handle("/api/", route())

	log.Println("server start")
	interfaceAddr, _ := net.InterfaceAddrs()
	for _, address := range interfaceAddr {
		ipNet, _ := address.(*net.IPNet)
		if ipNet.IP.To4() != nil {
			log.Printf("server addr http://%s:%s", ipNet.IP.String(), serverPort)
		}
	}

	err := http.ListenAndServe(":"+serverPort, nil)
	if err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}

func route() http.Handler {
	return &API{}
}

type API struct {
}

func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	apiMap[path](w, r)
}

var apiMap = map[string]func(w http.ResponseWriter, r *http.Request){
	"/api/chat/list": func(w http.ResponseWriter, r *http.Request) {
		// 聊天列表
		params := r.URL.Query()
		pageIndex, _ := strconv.Atoi(params["pageIndex"][0])
		pageSize, _ := strconv.Atoi(params["pageSize"][0])
		result, err := json.Marshal(enmicromsg.ChatList(pageIndex-1, pageSize))
		if err != nil {
			log.Fatalf("json marshal error: %v", err)
		}
		w.Write(result)
	},
	"/api/chat/detail": func(w http.ResponseWriter, r *http.Request) {
		//聊天记录
		params := r.URL.Query()
		talker := params["talker"][0]
		pageIndex, _ := strconv.Atoi(params["pageIndex"][0])
		pageSize, _ := strconv.Atoi(params["pageSize"][0])

		result, err := json.Marshal(enmicromsg.ChatDetailList(talker, pageIndex-1, pageSize))
		if err != nil {
			log.Fatalf("json marshal error: %v", err)
		}
		w.Write(result)
	},
	"/api/user/info": func(w http.ResponseWriter, r *http.Request) {
		// 用户信息
		params := r.URL.Query()
		username := params["username"][0]
		result, err := json.Marshal(enmicromsg.UserInfo(username))
		if err != nil {
			log.Fatalf("json marshal error: %v", err)
		}
		w.Write(result)
	},
	"/api/media/img": func(w http.ResponseWriter, r *http.Request) {
		// 图片
		params := r.URL.Query()
		msgId := params["msgId"][0]
		w.Write([]byte(wxfileindex.GetImgPath(msgId)))
	},
	"/api/media/video": func(w http.ResponseWriter, r *http.Request) {
		// TODO 视频
		params := r.URL.Query()
		msgId := params["msgId"][0]
		w.Write([]byte(wxfileindex.GetVideoPath(msgId)))
	},
	"/api/media/voice": func(w http.ResponseWriter, r *http.Request) {
		// TODO 语音
		params := r.URL.Query()
		msgId := params["msgId"][0]
		w.Write([]byte(wxfileindex.GetVoicePath(msgId)))
	},
}
