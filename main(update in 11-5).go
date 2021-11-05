package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	Regist  = fakeActive("regist")
	Payment = fakeActive("payment")
	Login   = fakeActive("login")
	//linux信号(用windows信号模拟了)
	sigChan chan os.Signal
)

type Server struct {
	name     string
	business string
}

//sever启动方法抽象
type Active func(ctx context.Context, business string) (Server, error)

//server启动方法实例
func fakeActive(kind string) Active {
	return func(ctx context.Context, business string) (Server, error) {
		s := Server{name: kind, business: business}
		err := s.Start(ctx)
		return s, err
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	//注册Linux系统信号,这里假设是interrupt，用windows信号模拟
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	Google := func(ctx context.Context, business string) ([]Server, error) {
		g, ctx := errgroup.WithContext(ctx)
		activees := []Active{Regist, Payment, Login}
		servers := make([]Server, len(activees))
		for i, active := range activees {
			i, active := i, active
			//用errgroup开goroutine，一个err全部shutdown
			g.Go(func() error {
				server, err := active(ctx, business)
				if err == nil {
					servers[i] = server
				}
				return err
			})
		}
		if err := g.Wait(); err != nil {
			return nil, err
		}
		return servers, nil
	}
	//5秒超时退出
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()
	servers, err := Google(ctx, "streaming")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println("-----List of available server:-----")
	for _, server := range servers {
		fmt.Printf("%v for %v", server.name, server.business)
		fmt.Println()
	}

}

//需要0~8秒启动Server
func (s *Server) Start(ctx context.Context) error {

	lucky := rand.Intn(8)

	if rand.Intn(3) == 0 {

		return fmt.Errorf("server %v start failed,stop activating", s.name)
	}

	for lucky > 0 {
		select {
		//接收ctx的结束信号
		case <-ctx.Done():
			return fmt.Errorf("time out,stop activating")
		//接收Linux系统信号
		case <-sigChan:
			return fmt.Errorf("linux interrupt signal,stop activating")
		//耗时启动
		default:
			fmt.Printf("activating%v in %vs\n", s.name, lucky)
			time.Sleep(time.Second)
			lucky--
		}
	}
	fmt.Printf("server %v started successfully!\n", s.name)
	return nil
}
