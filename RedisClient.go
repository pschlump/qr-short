package main

import (
	"fmt"
	"os"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/radix.v2/redis"
)

func RedisClient() (client *redis.Client, conFlag bool) {
	var err error
	fmt.Printf("AT: connect to redis with: %s %s\n", godebug.LF(), gCfg.RedisConnectHost+":"+gCfg.RedisConnectPort)
	client, err = redis.Dial("tcp", gCfg.RedisConnectHost+":"+gCfg.RedisConnectPort)
	if err != nil {
		fmt.Printf("Error on connect to redis:%s, fatal\n", err)
		fmt.Fprintf(os.Stderr, "%s\n\n\n-----------------------------------------------------------------------------------------------\nError on connect to redis:%s, fatal\n", MiscLib.ColorRed, err)
		fmt.Fprintf(os.Stderr, "Config Data: %s\n", godebug.SVarI(gCfg))
		fmt.Fprintf(os.Stderr, "\n-----------------------------------------------------------------------------------------------\n\n\n%s", MiscLib.ColorReset)
		os.Exit(1)
	}
	if gCfg.RedisConnectAuth != "" {
		err = client.Cmd("AUTH", gCfg.RedisConnectAuth).Err
		if err != nil {
			fmt.Printf("Error on connect to Redis --- Invalid authentication:%s, fatal\n", err)
			fmt.Fprintf(os.Stderr, "%s\nError on connect to Redis --- Invalid authentication:%s, fatal%s\n\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
			os.Exit(1)
		}
		conFlag = true
	}
	conFlag = true
	return
}
