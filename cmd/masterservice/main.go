package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	ds "github.com/zilliztech/milvus-distributed/internal/dataservice"
	dsc "github.com/zilliztech/milvus-distributed/internal/distributed/dataservice"
	isc "github.com/zilliztech/milvus-distributed/internal/distributed/indexservice/client"
	msc "github.com/zilliztech/milvus-distributed/internal/distributed/masterservice"
	psc "github.com/zilliztech/milvus-distributed/internal/distributed/proxyservice"
	is "github.com/zilliztech/milvus-distributed/internal/indexservice"
	ms "github.com/zilliztech/milvus-distributed/internal/masterservice"
	"github.com/zilliztech/milvus-distributed/internal/proto/commonpb"
	"github.com/zilliztech/milvus-distributed/internal/proto/internalpb2"
)

const reTryCnt = 3

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Printf("master service address : %s:%d", ms.Params.Address, ms.Params.Port)

	svr, err := msc.NewGrpcServer(ctx)
	if err != nil {
		panic(err)
	}

	log.Printf("proxy service address : %s", psc.Params.NetworkAddress())
	//proxyService := psc.NewClient(ctx, psc.Params.NetworkAddress())

	//TODO, test proxy service GetComponentStates, before set

	//if err = svr.SetProxyService(proxyService); err != nil {
	//	panic(err)
	//}

	log.Printf("data service address : %s:%d", ds.Params.Address, ds.Params.Port)
	dataService := dsc.NewClient(fmt.Sprintf("%s:%d", ds.Params.Address, ds.Params.Port))
	if err = dataService.Init(); err != nil {
		panic(err)
	}
	if err = dataService.Start(); err != nil {
		panic(err)
	}
	cnt := 0
	for cnt = 0; cnt < reTryCnt; cnt++ {
		dsStates, err := dataService.GetComponentStates()
		if err != nil {
			continue
		}
		if dsStates.Status.ErrorCode != commonpb.ErrorCode_SUCCESS {
			continue
		}
		if dsStates.State.StateCode != internalpb2.StateCode_INITIALIZING && dsStates.State.StateCode != internalpb2.StateCode_HEALTHY {
			continue
		}
		break
	}
	if cnt >= reTryCnt {
		panic("connect to data service failed")
	}

	//if err = svr.SetDataService(dataService); err != nil {
	//	panic(err)
	//}

	log.Printf("index service address : %s", is.Params.Address)
	indexService := isc.NewClient(is.Params.Address)

	if err = svr.SetIndexService(indexService); err != nil {
		panic(err)
	}

	if err = svr.Start(); err != nil {
		panic(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	sig := <-sc
	log.Printf("Got %s signal to exit", sig.String())
	_ = svr.Stop()
}