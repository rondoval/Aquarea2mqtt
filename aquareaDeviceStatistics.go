package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Gets most recent data from the statistics page
func (aq *aquarea) getDeviceLogInformation(user aquareaEndUserJSON, shiesuahruefutohkun string) (map[string]string, error) {
	// Build list of all possible values to log
	var valueList strings.Builder
	valueList.WriteString("{\"logItems\":[")
	for i := range aq.logItems {
		valueList.WriteString(strconv.Itoa(i))
		valueList.WriteString(",")
	}
	valueList.WriteString("]}")

	b, err := aq.httpPost(aq.AquareaServiceCloudURL+"/installer/api/data/log", url.Values{
		"var.deviceId":        {user.DeviceID},
		"shiesuahruefutohkun": {shiesuahruefutohkun},
		"var.target":          {"0"},
		"var.startDate":       {fmt.Sprintf("%d000", time.Now().Unix()-aq.logSecOffset)},
		"var.logItems":        {valueList.String()},
	})
	if err != nil {
		return nil, err
	}
	var aquareaLogData aquareaLogDataJSON
	err = json.Unmarshal(b, &aquareaLogData)
	if err != nil {
		return nil, err
	}

	var deviceLog map[int64][]string
	err = json.Unmarshal([]byte(aquareaLogData.LogData), &deviceLog)
	if err != nil {
		return nil, err
	}
	if len(deviceLog) < 1 {
		// no data in log
		return nil, nil
	}

	// we're interested in the most recent snapshot only
	var lastKey int64 = 0
	for k := range deviceLog {
		if lastKey < k {
			lastKey = k
		}
	}

	unitRegexp := regexp.MustCompile(`(.+)\[(.+)\]`)

	stats := make(map[string]string)
	for i, val := range deviceLog[lastKey] {
		split := unitRegexp.FindStringSubmatch(aq.logItems[i])

		topic := "log/" + strings.ReplaceAll(strings.Title(split[1]), " ", "")
		stats[topic+"/unit"] = split[2] // unit of the value, extracted from name
		stats[topic] = val
	}
	stats["log/Timestamp"] = strconv.FormatInt(lastKey, 10)
	stats["log/CurrentError"] = strconv.Itoa(aquareaLogData.ErrorCode)
	stats["EnduserID"] = user.Gwid
	return stats, nil
}