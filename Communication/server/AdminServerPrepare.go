package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

var wgOut sync.WaitGroup


func main() {
	//---------------------------------------------//

	wgOut.Add(6)
	routeMultiplexer:= calculateRoutingTable(6,"127.0.0.1",12341)
	fmt.Println()
	for route := range routeMultiplexer.routeAddresses{
		fmt.Println(route,routeMultiplexer.nodeAddresses[strconv.FormatUint(uint64(route),10)][2], routeMultiplexer.routeAddresses[route])
		if route==2147483646{
			continue
		}
		go creatServerInstance(
			routeMultiplexer.nodeAddresses[strconv.FormatUint(uint64(route),10)][2],
			route,
			routeMultiplexer.routeAddresses[route])
	}

	time.Sleep(time.Second*15)
	go creatFailedServerInstance(
		routeMultiplexer.nodeAddresses[strconv.FormatUint(uint64(2147483646),10)][2],
		"127.0.0.1",
		2147483646,
		routeMultiplexer.routeAddresses[2147483646])



	wgOut.Wait()

}


