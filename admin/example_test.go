package admin_test

import (
	"fmt"
	"io"
	"net/http"

	"github.com/moov-io/base/admin"
)

func Example() {
	svc, err := admin.New(admin.Opts{Addr: "127.0.0.1:0"})
	if err != nil {
		panic(err)
	}

	svc.AddLivenessCheck("db", func() error { return nil })
	svc.AddReadinessCheck("kafka", func() error {
		return fmt.Errorf("not connected")
	})
	svc.AddVersionHandler("v1.2.3")

	go svc.Listen()
	defer svc.Shutdown()

	base := "http://" + svc.BindAddr()

	printBody := func(path string) {
		resp, err := http.Get(base + path)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(path, resp.StatusCode, string(body))
	}

	printBody("/live")
	printBody("/ready")
	printBody("/version")
	// Output:
	// /live 200 {"db":"good"}
	// /ready 400 {"kafka":"not connected"}
	// /version 200 v1.2.3
}
