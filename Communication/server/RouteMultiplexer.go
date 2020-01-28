package main

import (
	"math"
	"strconv"
)

type RouteMultiplexer struct{
	interval uint32
	maxV uint32
	numberNode int
	nodes []uint32
	maxRouteVal int
	logValue int
	routingTable map[uint32][]string
	//....local use variables....//
	nodeAddresses map[string][]string
	routeAddresses map[uint32][][]string
}

func (multiplexer *RouteMultiplexer) setupVariables(numberOfNode int){
	multiplexer.numberNode=numberOfNode
	multiplexer.maxV=^uint32(0)
	multiplexer.maxRouteVal= int(math.Ceil(math.Log2(float64(multiplexer.numberNode))))
	multiplexer.interval=multiplexer.maxV/uint32(multiplexer.numberNode)
	multiplexer.logValue= int(math.Ceil(math.Log2(float64(multiplexer.interval))))
}

func (multiplexer *RouteMultiplexer) findAllNodeId(){
	for i:=1;i<=multiplexer.numberNode;i++{
		multiplexer.nodes=append(multiplexer.nodes,multiplexer.interval*uint32(i))
	}
}

func (multiplexer *RouteMultiplexer) calculateRouteEntries() map[uint32][]string {
	multiplexer.routingTable=make(map[uint32][]string)
	for _, node :=range multiplexer.nodes{
		currentRoutes :=[]string{}
		for i:=0; i<multiplexer.maxRouteVal;i++{
			current:=uint32(int64(float64(node)+ math.Pow(2,float64(multiplexer.logValue+i)))%int64(multiplexer.maxV))
			closeNode:=multiplexer.findClosestNode(current)
			if closeNode != node {currentRoutes =append(currentRoutes,strconv.FormatUint(uint64(closeNode),10))}
		}
		multiplexer.routingTable[node]= currentRoutes
	}
	return multiplexer.routingTable
}

func (multiplexer *RouteMultiplexer) findClosestNode(intValue uint32) uint32{
	bestChoice:= multiplexer.nodes[0]
	difference:= multiplexer.nodes[0]-intValue
	for _,node:= range multiplexer.nodes{
		if node-intValue >=0  && difference < node-intValue{
			difference=node-intValue
			bestChoice=node
		}
	}
	return bestChoice
}

// ..... LOCAL USE ONLY!!.......

func (multiplexer *RouteMultiplexer) setAddresses(address string, port int64) map[string][]string{
	multiplexer.nodeAddresses=make(map[string][]string)
	var portCounter int64=0
	for _,node:= range multiplexer.nodes{
		nodeStrVal:=strconv.FormatUint(uint64(node),10)
		values:=[]string{nodeStrVal,address, strconv.FormatInt(port+portCounter,10)}
		multiplexer.nodeAddresses[nodeStrVal]= values
		portCounter+=2
	}
	return  multiplexer.nodeAddresses
}

func (multiplexer *RouteMultiplexer)  getAllNodeAddresses(){
	multiplexer.routeAddresses=make(map[uint32][][]string)

	for route := range multiplexer.routingTable{
		routeAddress:=[][]string{}
		for _,node:= range multiplexer.routingTable[route]{
			routeAddress=append(routeAddress,multiplexer.nodeAddresses[node])
		}
		multiplexer.routeAddresses[route]=routeAddress;
	}
}

// ..... PREPARE MULTIPLEXER OBJECT ......

func calculateRoutingTable(numberOfNodes int, ip string, port int64)  RouteMultiplexer{
	routeMultiplexer:=RouteMultiplexer{}
	routeMultiplexer.setupVariables(numberOfNodes)
	routeMultiplexer.findAllNodeId()
	routeMultiplexer.calculateRouteEntries()
	//..... setting up addresses (local use only!)
	routeMultiplexer.setAddresses(ip,port)
	routeMultiplexer.getAllNodeAddresses()
	return routeMultiplexer
}
