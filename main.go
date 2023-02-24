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
	"time"
)

const API = "https://gitee.com/api/v5/repos/%s/%s/contents/%s"

var (
	token   string
	owner   string
	repo    string
	message string
	branch  string
	path    string
	paths   = []string{}
)

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
	for _, v := range paths {
		if exist, _ := fileExists(v); !exist {
			continue
		}
		content, err := os.ReadFile(v)
		if err != nil {
			fmt.Println("[Error] Read image file error, ", err.Error())
			continue
		}
		tmpData := url.Values{}
		if branch != "" {
			tmpData.Add("branch", branch)
		}
		tmpData.Add("access_token", token)
		tmpData.Add("message", message)
		tmpPath := path + "image_" + strconv.FormatInt(time.Now().Unix(), 10) + "_" + filepath.Base(v)
		postApi := fmt.Sprintf(API, owner, repo, tmpPath)
		tmpData.Add("content", base64.StdEncoding.EncodeToString(content))
		resp, err := http.PostForm(postApi, tmpData)
		if err != nil {
			fmt.Println("[Error] Upload image file error, ", err.Error())
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("[Error] Parse resp body error, ", err.Error())
			continue
		}
		if resp.StatusCode == 201 {
			fmt.Println("https://gitee.com/" + owner + "/" + repo + "/raw/master/" + tmpPath)
		} else {
			fmt.Println("[Error] Upload to gitee error, resp ↓\n" + string(body))
			fmt.Println(string(body))
		}
	}
}

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

func main() {
	bindFlags()
	upload()
}
