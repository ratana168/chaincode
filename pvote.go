package main

import (
	"fmt"
	"flag"
	"sync"
	"bytes"
	"net/http"
	"time"
//	"io/ioutil"
)

/*
 * ./pvote -s 1 -e 100 -p 10 -c 1001 -a a -url http://13.113.79.84:7050/chaincode -ccname xxxxxx
*/
func main() {
	var (
		prefix 		string
		startnum 	int64
		endnum		int64
		candidateid	string
		parallel	int
		ccname		string
		ipaddr		string
		ua			string
		url			string
		timeout		int
	)

	flag.StringVar(&prefix, "a", "", "prefix characters")
	flag.Int64Var(&startnum, "s", 0, "start number")
	flag.Int64Var(&endnum, "e", 0, "end number")
	flag.StringVar(&candidateid, "c", "1001", "candidate id to vote")
	flag.IntVar(&parallel, "p", 1, "number of parallel processes")
	flag.StringVar(&ccname, "ccname", "", "chaincode ID")
	flag.StringVar(&ipaddr, "ipaddr", "127.0.0.1", "IP address")
	flag.StringVar(&ua, "ua", "pvote", "user agent")
	flag.StringVar(&url, "url", "http://localhost:7050/chaincode", "url to call rest api")
	flag.IntVar(&timeout, "timeout", 10, "http connection timeout (seconds)")
	flag.Parse()

	n := endnum - startnum + 1
	q := n / int64(parallel)
	r := n % int64(parallel)

	wg := &sync.WaitGroup{}
	sp := startnum
	ep := startnum + q + r - 1
	for i := 1; i <= parallel; i++ {
		wg.Add(1)
		go func(s int64, e int64) {
			defer wg.Done()
			for t := s; t <= e; t++ {
				seq := fmt.Sprintf("%016d", t)
				token := prefix + seq[len(prefix):16]
				t := time.Now()
				datetime := t.Format("2006-01-02 15:04:05")
				data := []byte(`{"jsonrpc": "2.0", "method": "invoke", "params": { "type": 1,	"chaincodeID": { "name": "` + ccname + `"},	"ctorMsg": { "function": "vote", "args": [ "` + token + `", "` + candidateid + `", "` + datetime + `", "` + ipaddr + `", "` + ua +`"]}, "secureContext": "jim"}, "id": 1}`)
				client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
				client.Post( url , "application/json", bytes.NewBuffer(data))
//				resp, _ := client.Post( url , "application/json", bytes.NewBuffer(data))
//				body, _ := ioutil.ReadAll(resp.Body)
//				fmt.Println(string(body))
			}
		}(sp, ep)
		sp = ep + 1
		ep = ep + q
	}

	wg.Wait()

}
