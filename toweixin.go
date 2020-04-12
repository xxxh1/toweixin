package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	sendurl   = `https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=`
	get_token = `https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=`
)

type access_token struct {
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
}

type send_msg_error struct {
	Errcode int    `json:"errcode`
	Errmsg  string `json:"errmsg"`
}

type send_msg struct {
	Touser  string            `json:"touser"`
	Toparty string            `json:"toparty"`
	Totag   string            `json:"totag"`
	Msgtype string            `json:"msgtype"`
	Agentid int               `json:"agentid"`
	Text    map[string]string `json:"text"`
	Safe    int               `json:"safe"`
}

var requestError = errors.New("request error,check url or network")

func main() {
	var contentlists []string
	agentid := flag.Int("i", 0, "-i 企业微信应用AgentId")
	content := flag.String("c", "", "-c '' 指定要发送的内容")
	corpid := flag.String("p", "", "-p 企业ID")
	corpsecret := flag.String("s", "", "-s 企业微信应用Secret")
	flag.Parse()
	sc := bufio.NewScanner(os.Stdin)
	if *content != "" {
		contentlists = []string{*content}

	} else {
		for sc.Scan() {
			contentlists = append(contentlists, sc.Text())
		}
	}
	if err := sc.Err(); err != nil {
		contentlists = []string{"骚年设置扫描器的报警信息"}
		fmt.Fprintf(os.Stderr, " failed to read input: %s\n ", err)
	}
	var content1 string
	for _, content := range contentlists {
		content1 = content1 + "\n" + content
	}
	var meg send_msg = send_msg{Touser: "@all", Msgtype: "text", Agentid: *agentid, Text: map[string]string{"content": content1}}

	token, err := Get_token(*corpid, *corpsecret)
	if err != nil {
		println(err.Error())
		return
	}
	buf, err := json.Marshal(meg)
	if err != nil {
		return
	}
	err = json.Unmarshal(buf, &meg)
	if err != nil {
		println(err)
		return
	}
	err = Send_msg(token.Access_token, buf)
	if err != nil {
		println(err.Error())
	}

}
func Get_token(corpid, corpsecret string) (at access_token, err error) {
	resp, err := http.Get(get_token + corpid + "&corpsecret=" + corpsecret)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = requestError
		return
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(buf, &at)
	if at.Access_token == "" {
		err = errors.New("corpid or corpsecret error.")
	}
	return
}

func Send_msg(Access_token string, msgbody []byte) error {
	body := bytes.NewBuffer(msgbody)
	resp, err := http.Post(sendurl+Access_token, "application/json", body)
	if resp.StatusCode != 200 {
		return requestError
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var e send_msg_error
	err = json.Unmarshal(buf, &e)
	if err != nil {
		return err
	}
	if e.Errcode != 0 && e.Errmsg != "ok" {
		return errors.New(string(buf))
	}
	return nil
}
