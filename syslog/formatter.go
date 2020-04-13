package syslog

import (
	"fmt"
)

const origLocalFmt = "<%d>%s %s[%d]: %s%s"
const origRemoteFmt = "<%d>%s %s %s[%d]: %s%s"
const datadogFmt = "<%s> <%d> %s %s %s - - - %s%s"

var localFmt = origLocalFmt
var remoteFmt = origRemoteFmt
var datadogKey = ""

func formatLogMsg(p Priority, timestamp, hostname, tag string, pid int, msg, nl string, isLocal bool) []byte {
	var result string
	if len(datadogKey) != 0 {
		//"<DATADOG_API_KEY> <%pri%>%protocol-version% %timestamp:::date-rfc3339% %HOSTNAME% %app-name% - - - %msg%\n"
		result = fmt.Sprintf(datadogFmt, datadogKey, p, timestamp, hostname, tag, msg, nl)
	} else if isLocal {
		result = fmt.Sprintf(localFmt, p, timestamp, tag, pid, msg, nl)
	} else {
		result = fmt.Sprintf(remoteFmt, p, timestamp, hostname, tag, pid, msg, nl)
	}

	return []byte(result)
}

func SetDataDogKey(k string) {
	datadogKey = k
}
