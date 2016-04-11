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
	"sync"
	"fmt"
)

var token = "388df3d7442cf9fd6c5cc9b43b9a015a247db04dc62f07c7aedc4d2fe9771067434bc5b1ebb503e57911e"


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

	sR := "https://api.vk.com/method/photos.getWallUploadServer?group_id=72507356&access_token="+token

	client := &http.Client{}

	r, _ := client.Get(sR)
	b, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal([]byte(b), &getWallUploadServers)
	if err != nil {
		log.Fatalf("uplServ error: %v", err)
	}else {
		fh, err := os.Open(img)
		if err != nil {
			log.Print("error opening file")
		}

		s,p,h,r := postImg(getWallUploadServers.Response.UploadURL,fh)

		ids := saveImg(s,p,h,r)

		postWall(id,message,ids,date)
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

	req0, err := http.NewRequest("POST", url, &b)
	if err != nil {
		log.Fatalf("NewRequest error: %v",err)
	}

	req0.Header.Set("Content-Type", w.FormDataContentType())

	client0 := &http.Client{}

	res0, err_postImg := client0.Do(req0)
	if err_postImg != nil {
		log.Fatalf("Do error: %v", err_postImg)
	}
	uplRes := UploadResponse{}

	dec := json.NewDecoder(res0.Body)
	err = dec.Decode(&uplRes)
	if err != nil {
		log.Fatalf("Decode error: %v", err)
	}
	defer res0.Body.Close()

	log.Print(uplRes.Server)

	server = uplRes.Server
	photo = uplRes.Photo
	hash = uplRes.Hash
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

	uplRes1 := saveWallPhoto{}

	uplErr  := Error{}

	data := url.Values{}
	data.Set("server", strconv.Itoa(server))
	data.Add("photo", photo)
	data.Add("hash", hash)
	data.Add("group_id", "72507356")

	url := "https://api.vk.com/method/photos.saveWallPhoto?access_token="+token

	clientp := &http.Client{}

	res1, _ := clientp.Post(url,"application/x-www-form-urlencoded",bytes.NewBufferString(data.Encode()))

	dec := json.NewDecoder(res1.Body)
	err = dec.Decode(&uplRes1)
	if err != nil {
		err = dec.Decode(&uplErr)
		if err != nil{
			log.Fatalf("Decode uplErr error: %v",err)
		}
		log.Fatalf("Decode1 error: %v",uplErr.Error.ErrorMsg)
	}

	log.Print(uplRes1.Response[0].ID)

	id = uplRes1.Response[0].ID

	defer res1.Body.Close()
	return id
}

func postWall(id int, message string, attachments string, date string)  {

	vkontakteUserId := "-72507356"

	client1 := &http.Client{}

	data := url.Values{}
	data.Set("owner_id", vkontakteUserId)
	data.Add("message", message)
	data.Add("attachments", attachments)
	data.Add("publish_date",date)

	log.Print("tut3")

	res2, err := client1.Post("https://api.vk.com/method/wall.post?access_token="+token,"application/x-www-form-urlencoded",bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Fatalf("postWall error: %v", err)
	}

	uplpostWall := postWalls{}

	dec := json.NewDecoder(res2.Body)

	err = dec.Decode(&uplpostWall)
	if err != nil {
		log.Fatalf("Decode postWall error: %v", err)
	}

	defer res2.Body.Close()

	//dbPost(id,uplpostWall.Response.PostID)

	log.Print("tut7")
}

func dbPost(id int,post_id int) {

	log.Print("tut4")

	db, err := sql.Open("mysql", "root:Vbirfufvvb@(127.0.0.1:3306)/vk")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	log.Print("tut5")

	stmt, err := db.Prepare("UPDATE queue SET post_id = ?, post=1 WHERE id = ?;")

	res, err := stmt.Exec(post_id,id)
	panic(err.Error())

	affect, err := res.RowsAffected()
	panic(err.Error())

	log.Print("tut6")

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

	//uplServ("1","./1.jpg","now")

	jsonResponses := make(chan string)

	joobs := getDb().Items

	var wg sync.WaitGroup

	wg.Add(len(joobs))

	log.Print(len(joobs))

	for _, joob := range joobs {
		go func(joob getdbs) {
			defer wg.Done()
			uplServ(joob.Id,joob.Message,joob.Image,joob.Date)
			jsonResponses <- string(joob.Message)
		}(joob)
	}

	go func() {
		for response := range jsonResponses {
			fmt.Println(response)
		}
	}()

	wg.Wait()
}
