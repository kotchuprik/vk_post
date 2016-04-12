package main

import (
	"net/http"
	"io/ioutil"
	"log"
	"encoding/json"
	"os"
	"mime/multipart"
	"bytes"
	"io"
	"strconv"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
	"fmt"
	"time"
	"math/rand"
)

var token = "388df3d7442cf9fd6c5cc9b43b9a015a247db04dc62f07c7aedc4d2fe9771067434bc5b1ebb503e57911e"

var group_id = 72507356

type getdbs struct {
	Id      int
	Message string
	Image	string
	Date	string
}

type boxs struct {
	Items []getdbs
}

type postWalls struct {
	Response struct {
			 PostID int `json:"post_id"`
		 } `json:"response"`
}

func (box *boxs) AddItem(item getdbs) []getdbs {
	box.Items = append(box.Items, item)
	return box.Items
}

func uplServ(id int,message string, img string, date string){
	type getWallUploadServer struct {
		Response struct {
				 Aid       int    `json:"aid"`
				 Mid       int    `json:"mid"`
				 UploadURL string `json:"upload_url"`
			 } `json:"response"`
	}

	var getWallUploadServers *getWallUploadServer

	r_uplServ := "https://api.vk.com/method/photos.getWallUploadServer?group_id=72507356&access_token="+token

	client := &http.Client{}

	res_uplServ, _ := client.Get(r_uplServ)
	b_uplServ, err := ioutil.ReadAll(res_uplServ.Body)
	if err != nil {
		log.Fatalf("b_uplServ error: %v", err)
	} else {
		err = json.Unmarshal([]byte(b_uplServ), &getWallUploadServers)
		if err != nil {
			log.Fatalf("uplServ error: %v", err)
		}else {
			defer res_uplServ.Body.Close()

			fh, err := os.Open(img)
			if err != nil {
				log.Print("error opening file")
			}
			defer fh.Close()

			s,p,h,r := postImg(getWallUploadServers.Response.UploadURL,fh)
			ids := saveImg(s,p,h,r)
			postWall(id,message,ids,date)
		}
	}
}

func postImg(url string, img io.Reader) (server int, photo, hash string, err error) {

	type UploadResponse struct {
		Server int    `json:"server"`
		Photo  string `json:"photo"`
		Hash   string `json:"hash"`
	}

	var b bytes.Buffer

	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("photo", "photo.jpg")
	if err != nil {
		log.Fatalf("CreateFormFile error: %v", err)
	}
	if _, err = io.Copy(fw, img); err != nil {
		log.Fatalf("Copy error: %v",err)
	}
	w.Close()

	req_postImg, err := http.NewRequest("POST", url, &b)
	if err != nil {
		log.Fatalf("NewRequest error: %v",err)
	}else {

		req_postImg.Header.Set("Content-Type", w.FormDataContentType())

		client_postImg := &http.Client{}

		res_postImg, err_postImg := client_postImg.Do(req_postImg)
		if err_postImg != nil {
			log.Fatalf("Do error: %v", err_postImg)
		}else {
			defer res_postImg.Body.Close()

			uplRes := UploadResponse{}

			dec := json.NewDecoder(res_postImg.Body)
			err = dec.Decode(&uplRes)
			if err != nil {
				log.Fatalf("Decode error: %v", err)
			} else {
				server = uplRes.Server
				photo = uplRes.Photo
				hash = uplRes.Hash
			}
		}
	}
	return
}

func saveImg(server int, photo, hash string, err error) (id string){

	type saveWallPhoto struct {
		Response []struct {
			Aid      int    `json:"aid"`
			Created  int    `json:"created"`
			Height   int    `json:"height"`
			ID       string `json:"id"`
			OwnerID  int    `json:"owner_id"`
			Pid      int    `json:"pid"`
			Src      string `json:"src"`
			SrcBig   string `json:"src_big"`
			SrcSmall string `json:"src_small"`
			Text     string `json:"text"`
			Width    int    `json:"width"`
		} `json:"response"`
	}

	type Error struct {
		Error struct {
			      ErrorCode     int    `json:"error_code"`
			      ErrorMsg      string `json:"error_msg"`
			      RequestParams []struct {
				      Key   string `json:"key"`
				      Value string `json:"value"`
			      } `json:"request_params"`
		      } `json:"error"`
	}

	upl_saveImg := saveWallPhoto{}

	uplErr  := Error{}

	data := url.Values{}
	data.Set("server", strconv.Itoa(server))
	data.Add("photo", photo)
	data.Add("hash", hash)
	data.Add("group_id", strconv.Itoa(group_id))

	url_saveImg := "https://api.vk.com/method/photos.saveWallPhoto?access_token="+token

	client_saveImg := &http.Client{}

	time.Sleep(2*time.Second)

	res_saveImg, err := client_saveImg.Post(url_saveImg,"application/x-www-form-urlencoded",bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}else {
		defer res_saveImg.Body.Close()

		dec := json.NewDecoder(res_saveImg.Body)
		err = dec.Decode(&upl_saveImg)
		if err != nil {
			err = dec.Decode(&uplErr)
			if err != nil{
				log.Fatalf("Decode uplErr error: %v",err)
			}else {
				log.Fatalf("Decode1 error: %v",uplErr.Error.ErrorMsg)
			}
		} else {
			id = upl_saveImg.Response[0].ID
		}
	}
	return
}

func postWall(id int, message string, attachments string, date string)  {

	client_postWall := &http.Client{}

	data := url.Values{}
	data.Set("owner_id", strconv.Itoa(group_id*-1))
	data.Add("message", message)
	data.Add("attachments", attachments)
	data.Add("publish_date",date)

	res_postWall, err := client_postWall.Post("https://api.vk.com/method/wall.post?access_token="+token,"application/x-www-form-urlencoded",bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Fatalf("postWall error: %v", err)
	} else {
		defer res_postWall.Body.Close()

		uplpostWall := postWalls{}

		dec := json.NewDecoder(res_postWall.Body)

		err = dec.Decode(&uplpostWall)
		if err != nil {
			log.Fatalf("Decode postWall error: %v", err)
		}else {
			log.Print(uplpostWall.Response.PostID)
		}

		//dbPost(id,uplpostWall.Response.PostID)
	}
}
func dbPost(id int,post_id int) {
	db, err := sql.Open("mysql", "root:Vbirfufvvb@(127.0.0.1:3306)/vk")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE queue SET post_id = ?, post=1 WHERE id = ?;")

	res, err := stmt.Exec(post_id,id)
	panic(err.Error())

	affect, err := res.RowsAffected()
	panic(err.Error())

	fmt.Println(affect)
}

func getDb() boxs {
	items := []getdbs{}
	box := boxs{items}

	db, err := sql.Open("mysql", "root:Vbirfufvvb@(127.0.0.1:3306)/vk")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	stmtOut, err := db.Query("SELECT id, message, image, unix_timestamp(date) as date, post FROM queue WHERE post = 0")
	if err != nil {
		panic(err.Error())
	}
	defer stmtOut.Close()

	for stmtOut.Next(){
		var id int
		var message string
		var image string
		var date string
		var post string

		err = stmtOut.Scan(&id, &message, &image, &date, &post)
		if err != nil {
			log.Print("getDb")
			panic(err.Error())
		}else {
			item := getdbs{Id:id,Message:message,Image:image,Date:date}
			box.AddItem(item)
		}
	}
	return box
}

func main() {
	joobs := getDb().Items

	log.Print(len(joobs))

	for _, joob := range joobs {
		time.Sleep(2*time.Second)
		uplServ(joob.Id,joob.Message,joob.Image,joob.Date)
	}
}



