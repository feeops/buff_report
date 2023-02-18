package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-version"

	"google.golang.org/grpc/credentials"

	"golang.org/x/net/context"
	// 导入grpc包 .
	"google.golang.org/grpc"
	// 导入刚才我们生成的代码所在的proto包 .
	pb "buff_report/proto"
)

var (
	ProductName    = "buff_report"
	ProductVersion = "1.3"
)

func auth(buffID string) error {
	// 连接grpc服务器
	config := &tls.Config{
		InsecureSkipVerify: false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	conn, err := grpc.DialContext(ctx,
		"gw.bufftools.com:443",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewTLS(config)))

	if err != nil {
		return fmt.Errorf("无法连接服务器:%w", err)
	}

	defer conn.Close()

	// 初始化Greeter服务客户端
	c := pb.NewAccountClient(conn)

	// 调用SayHello接口，发送一条消息
	req := new(pb.Request)
	req.ProductName = ProductName
	req.ProductVersion = ProductVersion
	req.BuffID = buffID

	r, err := c.Auth(ctx, req)
	if err != nil {
		return fmt.Errorf("验证失败::%w", err)
	}

	logger.Info().Str("ExpireTime", r.ExpireTime).Str("AccountID", r.AccountID).
		Str("ProductVersion", r.ProductVersion).Msg("response")

	os.WriteFile("账号信息.txt", []byte(r.AccountID), 0666)

	URL := fmt.Sprintf("[InternetShortcut]\r\nURL=https://www.bufftools.com/buy?accountID=%s",
		r.AccountID)
	os.WriteFile("续费链接.url", []byte(URL), 0666)

	fmt.Printf("当前软件版本%s 最新软件版本%s\n", ProductVersion, r.ProductVersion)
	fmt.Printf("当前账号ID:%s 过期时间:%s\n", r.AccountID, r.ExpireTime)
	os.Stdout.Sync()

	v1, _ := version.NewVersion(ProductVersion)
	v2, _ := version.NewVersion(r.ProductVersion)

	if v1.LessThan(v2) {
		return fmt.Errorf("当前版本过老，请去https://www.bufftools.com上下载最新版本")
	}

	if r.ExpireSec <= 0 {
		return fmt.Errorf("试用期：%s已过，请及时续费", r.ExpireTime)
	}

	return nil
}
