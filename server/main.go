package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"
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
const MXCNT=2000
var fileCh = make(chan revdata,1000000)
func rev(res http.ResponseWriter,req *http.Request){
	defer req.Body.Close()
	var val=req.PostFormValue("content")
	var username=req.PostFormValue("username")
	fileCh<-revdata{username,val}
	res.WriteHeader(200)
	res.Write([]byte("ok"))
}
func send(res http.ResponseWriter,req *http.Request){
	defer req.Body.Close()
	res.WriteHeader(200)
	if bs,err:=ioutil.ReadFile("save.json");err==nil{
		var tmpfilecontent=filecontent{}
		if merr:=json.Unmarshal(bs,&tmpfilecontent);merr==nil{
			tmpfilecontent.Contents=append(tmpfilecontent.Contents,fcont.Contents...)
			if writedata,writeerr:=json.Marshal(tmpfilecontent);writeerr==nil{
				res.Write(writedata)
			}
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
		}
	}
}
func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/Submit",rev)
	http.HandleFunc("GetData",send)
	go writeFile()
	http.ListenAndServe(":1234",nil)
}