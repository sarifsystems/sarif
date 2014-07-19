package client

import (
	"crypto/rand"
)

func GenerateId() string {
	const alphanum = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	var bytes = make([]byte, 8)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func GetTopic(action, device, domain string) string {
	t := "stark"
	if device != "" {
		t += "/dev/" + device
	} else if domain != "" {
		t += "/domain/" + domain
	} else {
		t += "/special/all"
	}
	if action != "" {
		t += "/action/" + action
	}
	return t
}
