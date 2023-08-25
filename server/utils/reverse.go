package utils

import "server/model/response"

func ReverseKline(slice []response.Kline) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}
