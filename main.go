package main

import (
	"crypto/md5"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
)

func main() {
	cli := newClient()
	cli.Uin = 2366039635
	cli.PasswordMd5 = md5.Sum([]byte("czhaohao51**"))
	resp, err := cli.Login()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp)
}

func newClient() *client.QQClient {
	c := client.NewClientEmpty()
	return c
}
