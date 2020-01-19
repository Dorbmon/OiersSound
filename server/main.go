package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"compress/gzip"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
	"bytes"
	"encoding/base64"
)
type data struct {
	Name string `json:"name"`
	Content string `json:"content"`
	Time string `json:"time"`
	Num int `json:"num"`
}
type revdata struct {
	Name string
	Val string
}
type filecontent struct {
	Contents []data `json:"contents"`
}
const MXCNT=1
var fmutex sync.RWMutex
var fileCh = make(chan revdata,1000000)
var IpLastVis map[string]time.Time
const (
	MaxTimeBetween = time.Minute
)
func rev(res http.ResponseWriter,req *http.Request){
	ip := req.RemoteAddr
	if _,ok := IpLastVis [ip];ok {
		if time.Now().Sub(IpLastVis[ip]) < MaxTimeBetween {
			res.Write([]byte("操作过于频繁！请稍后再试"))
			return
		}
	}
	IpLastVis [ip] = time.Now()
	res.Header().Set("Access-Control-Allow-Origin","*")
	res.Header().Add("Access-Control-Allow-Headers","Content-Type")
	res.Header().Set("content-type","text/plain")
	defer req.Body.Close()
	var val=req.PostFormValue("content")
	var username=req.PostFormValue("username")
	fileCh<-revdata{username,val}
	res.WriteHeader(200)
	res.Write([]byte("成功！请刷新页面"))
}
func send(res http.ResponseWriter,req *http.Request){
	res.Header().Set("Access-Control-Allow-Origin","*")
	res.Header().Add("Access-Control-Allow-Headers","Content-Type")
	res.Header().Set("content-type","text/plain")
	defer req.Body.Close()
	res.WriteHeader(200)
	if bs,err:=ioutil.ReadFile("save.json");err==nil{
		var tmpfilecontent=filecontent{}
		if merr:=json.Unmarshal(bs,&tmpfilecontent);merr==nil{
			fmutex.RLock()
			defer fmutex.RUnlock()
			data,_ := ioutil.ReadFile("save.json")
			var buff bytes.Buffer
			gz := gzip.NewWriter(&buff)
			gz.Write(data)
			gz.Flush()
			gz.Close()
			res.Write([]byte(base64.StdEncoding.EncodeToString(buff.Bytes())))
		}else{
			fmt.Println("no no no")
		}
	}
}
var fcont filecontent
func writeFile(){
	var cnt = 0
	var gidx=1
	for{
		if sval,ok:=<-fileCh;ok{
			fmutex.Lock()
			var tdata=data{}
			tdata.Num=gidx
			gidx++
			tdata.Time=time.Now().String()
			tdata.Content=sval.Val
			tdata.Name=sval.Name
			fcont.Contents=append(fcont.Contents,tdata)
			cnt++
			if cnt>=MXCNT{
				if fdata,ferr:=ioutil.ReadFile("save.json");ferr==nil{
					var tmpfilecontent=filecontent{}
					if merr:=json.Unmarshal(fdata,&tmpfilecontent);merr==nil{
						tmpfilecontent.Contents=append(tmpfilecontent.Contents, fcont.Contents...)
						if writedata,writeerr:=json.Marshal(tmpfilecontent);writeerr==nil{
							os.Remove("save.json")
							ioutil.WriteFile("save.json",writedata,os.ModeAppend)
						}
					}else{
						fmt.Println("no no no")
					}
				}else{
					fmt.Println("no no no")
				}
				fcont.Contents=nil
				cnt=0

			}
			fmutex.Unlock()
		}
	}
}
func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	IpLastVis = make (map[string]time.Time)
	http.HandleFunc("/Submit",rev)
	http.HandleFunc("/GetData",send)
	go writeFile()
	fmt.Println("run run")
	http.ListenAndServe("0.0.0.0:1234",nil)
}