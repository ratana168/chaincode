package main

import (
	"fmt"
	"net/http"
	"time"
	"bytes"
	"io/ioutil"
)

const ccname = "6069c2740bab37d4f060b4f5d0d2e19f111b9bf31d7bfcde0803cd31f8c6f2f0a5eb2a67799be69cb892ac12f38c1ebaa4f7ff0cfbe2354f6f827fb97f52b3dc"

func main() {
	http.HandleFunc("/", top)
	http.HandleFunc("/vote", vote)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func top(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
<!DOCTYPE html PUBLIC "-//W3C//DTD Compact HTML 1.0 Draft//EN">
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
<meta http-equiv="Content-Style-Type" content="text/css">
<meta name="viewport" content="width=device-width, user-scalable=yes, initial-scale=1, maximum-scale=10">
<title>Vote Web Top Page</title>
<style type="text/css">
<!--
a:link {color: #FF3366;}
a:visited {color: #FF3366;}
-->
</style>
</head>
<body bgcolor="#FFFFFF" text="#000000" link="#FF3366" vlink="#FF3366">
<h2>Vote Web Top Page</font><h2><br>
<form method="POST" action="/vote">
<div>Vote Token: <input type="text" name="token"></div>
<div>Candidate: 
<select name="candidateid">
<option value="1001" selected>Hanako Suzuki</option>
<option value="1002">Midori Yamamoto</option>
<option value="1003">Keiko Tanaka</option>
</select>
</div>
<div><input type='submit' value='Vote'></div>
</form>
</body>
</html>
`)
}

func vote(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	token := r.FormValue("token")
	candidateid := r.PostFormValue("candidateid")
	t := time.Now()
	datetime := t.Format("2006-01-02 15:04:05")
	ipaddr := r.Header.Get("X-Forwarded-For")
	ua := r.UserAgent()
	data := []byte(`{"jsonrpc": "2.0", "method": "invoke", "params": { "type": 1,	"chaincodeID": { "name": "` + ccname + `"},	"ctorMsg": { "function": "vote", "args": [ "` + token + `", "` + candidateid + `", "` + datetime + `", "` + ipaddr + `", "` + ua +`"]}, "secureContext": "jim"}, "id": 2}`)
	resp, _ := client.Post( "http://localhost:7050/chaincode" , "application/json", bytes.NewBuffer(data))
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Fprintf(w, string(body))

}
