package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// gitee上传文件的api
const API = "https://gitee.com/api/v5/repos/%s/%s/contents/%s"

var (
	token   string       // gitee的token
	owner   string       // gitee的owner
	repo    string       // gitee的repo
	message string       // 上传时commit的message
	branch  string       // 上传到repo的分支
	path    string       // 上传到repo的路径
	paths   = []string{} // 本地图片路径或互联网图片路径
)

// bindFlags 绑定参数
func bindFlags() {
	flag.StringVar(&token, "token", "", "gitee token")
	flag.StringVar(&owner, "owner", "", "gitee owner")
	flag.StringVar(&repo, "repo", "", "gitee repo")
	flag.StringVar(&branch, "branch", "", "gitee repo branch")
	flag.StringVar(&path, "path", "/", "gitee repo path")
	flag.StringVar(&message, "message", "upload image", "gitee upload action message")

	flag.Usage = func() {
		fmt.Println("Usage: gitee_image_upload [-h] [-token string] [-owner string] [-repo string] [-path string] [-branch string] [-message string] {image_path}")
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("Usage: gitee_image_upload [-h] [-token string] [-owner string] [-repo string] [-path string] [-branch string] [-message string] {image_path}")
		flag.PrintDefaults()
		os.Exit(0)
	}
	for i := 0; i < flag.NArg(); i++ {
		paths = append(paths, flag.Arg(i))
	}
}

// upload 上传图片到gitee的总执行流程
func upload() {
	if token == "" {
		fmt.Println("[Error] Empty token")
		return
	}
	if owner == "" {
		fmt.Println("[Error] Empty owner")
		return
	}
	if repo == "" {
		fmt.Println("[Error] Empty repo")
		return
	}
	fmt.Println("[Info] Start uploading")
	postData := url.Values{} // postData
	postData.Add("access_token", token)
	postData.Add("message", message)
	if branch != "" {
		postData.Add("branch", branch)
	}
	for _, v := range paths {
		uploadImage(v, postData)
	}
}

// uploadImage 上传图片到gitee
func uploadImage(imagePath string, postData url.Values) {
	var (
		filename  string
		imageData []byte
		resp      *http.Response
		err       error
	)

	if isHttpImage(imagePath) {
		filename = filepath.Base(imagePath)
		resp, err = http.Get(imagePath)
		if err != nil {
			fmt.Println("[Error] Download http image file error, ", err.Error())
			return
		}
		defer resp.Body.Close()
		imageData, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("[Error] Read image file error, ", err.Error())
			return
		}
	} else if exist, _ := fileExists(imagePath); exist {
		filename = filepath.Base(imagePath)
		imageData, err = os.ReadFile(imagePath)
		if err != nil {
			fmt.Println("[Error] Read image file error, ", err.Error())
			return
		}
	} else {
		fmt.Println("[Error] Image file path error")
		return
	}
	uploadPath := path + "image_" + strconv.FormatInt(time.Now().Unix(), 10) + "_" + filename
	postApi := fmt.Sprintf(API, owner, repo, uploadPath)
	postData.Add("content", base64.StdEncoding.EncodeToString(imageData))
	resp, err = http.PostForm(postApi, postData)
	if err != nil {
		fmt.Println("[Error] Upload image file error, ", err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[Error] Parse resp body error, ", err.Error())
		return
	}
	if resp.StatusCode == 201 {
		fmt.Println("https://gitee.com/" + owner + "/" + repo + "/raw/master/" + uploadPath)
	} else {
		fmt.Println("[Error] Upload to gitee error, resp ↓\n" + string(body))
	}
}

// isHttpImage 检测是否是url图片资源
func isHttpImage(path string) bool {
	protols := [...]string{"https://", "http://"}
	types := [...]string{"png", "jpg", "jpeg", "gif"}
	for _, v := range protols {
		if strings.HasPrefix(path, strings.ToLower(v)) {
			for _, v := range types {
				if strings.HasSuffix(path, strings.ToLower(v)) {
					return true
				}
			}
		}
	}
	return false
}

// fileExists 检测文件是否存在
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// main 主程序
func main() {
	bindFlags()
	upload()
}
