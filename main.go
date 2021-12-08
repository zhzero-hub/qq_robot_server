package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"

	"bytes"
	"time"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/go-cqhttp/coolq"
	"github.com/Mrs4s/go-cqhttp/global"
	"github.com/Mrs4s/go-cqhttp/modules/servers"
	"github.com/gocq/qrcode"
	log "github.com/sirupsen/logrus"
)

var cli *client.QQClient

func main() {
	cli = newClient()
	cli.Uin = 2366039635
	cli.PasswordMd5 = md5.Sum([]byte("czhaohao51**"))
	if token, err := os.ReadFile("session.token"); err == nil {
		err = cli.TokenLogin(token)
		if err != nil {
			fmt.Println(err)
			_ = os.Remove("session.token")
			return
		}
	} else {
		fmt.Println("普通登录")
		resp, err := cli.Login()
		if err != nil {
			fmt.Println(err)
			return
		} else if !resp.Success {
			fmt.Println("普通登录失败，开始二维码登录")
			err = qrcodeLogin()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	fmt.Println("登录成功")
	token := cli.GenToken()
	_ = os.WriteFile("session.token", token, 0644)
	byre, _ := json.Marshal(cli)
	fmt.Println(string(byre))
	cli.ReloadFriendList()
	log.Infof("共加载 %v 个好友.", len(cli.FriendList))
	cli.ReloadGroupList()
	log.Infof("共加载 %v 个群.", len(cli.GroupList))
	cli.SetOnlineStatus(0)
	servers.Run(coolq.NewQQBot(cli))
	<-global.SetupMainSignalHandler()
}

func newClient() *client.QQClient {
	c := client.NewClientEmpty()
	return c
}

func qrcodeLogin() error {
	fmt.Println("获取二维码")
	rsp, err := cli.FetchQRCode()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fi, err := qrcode.Decode(bytes.NewReader(rsp.ImageData))
	if err != nil {
		return err
	}
	_ = os.WriteFile("qrcode.png", rsp.ImageData, 0o644)
	defer func() { _ = os.Remove("qrcode.png") }()
	if cli.Uin != 0 {
		log.Infof("请使用账号 %v 登录手机QQ扫描二维码 (qrcode.png) : ", cli.Uin)
	} else {
		log.Infof("请使用手机QQ扫描二维码 (qrcode.png) : ")
	}
	time.Sleep(time.Second)
	qrcodeTerminal.New2(qrcodeTerminal.ConsoleColors.BrightBlack, qrcodeTerminal.ConsoleColors.BrightWhite, qrcodeTerminal.QRCodeRecoveryLevels.Low).Get(fi.Content).Print()
	s, err := cli.QueryQRCodeStatus(rsp.Sig)
	if err != nil {
		return err
	}
	prevState := s.State
	for {
		time.Sleep(time.Second)
		s, _ = cli.QueryQRCodeStatus(rsp.Sig)
		if s == nil {
			continue
		}
		if prevState == s.State {
			continue
		}
		prevState = s.State
		switch s.State {
		case client.QRCodeCanceled:
			log.Fatalf("扫码被用户取消.")
		case client.QRCodeTimeout:
			log.Fatalf("二维码过期")
		case client.QRCodeWaitingForConfirm:
			log.Infof("扫码成功, 请在手机端确认登录.")
		case client.QRCodeConfirmed:
			log.Infof("扫码成功，开始登录")
			res, err := cli.QRCodeLogin(s.LoginInfo)
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println(res)
			return loginResponseProcessor(res)
		case client.QRCodeImageFetch, client.QRCodeWaitingForScan:
			// ignore
		}
	}
}

func loginResponseProcessor(res *client.LoginResponse) error {
	if res.Success {
		return nil
	} else {
		fmt.Println(res.Error)
		return nil
	}
}
