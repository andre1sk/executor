package main

import "fmt"

var Actions = map[string]func(in string) (out string, err error){
	"email_get_body":       EmailBodyGet,
	"str_extract_url":      StrExtractURL,
	"url_check_reputation": URLCheckReputation,
	"dummy":                DummyAction,
}

func EmailBodyGet(in string) (out string, err error) {
	out = "Email body"
	return
}

func StrExtractURL(in string) (out string, err error) {
	out = "http://example.com"
	return
}

func URLCheckReputation(in string) (out string, err error) {
	out = "Good"
	return
}

func DummyAction(in string) (out string, err error) {
	out = "Dummy"
	return
}

func RunAction(action string, in string) (out string, err error) {
	if _, ok := Actions[action]; !ok {
		return "", fmt.Errorf("action %s not found", action)
	}
	return Actions[action](in)
}

func ActionExists(action string) bool {
	_, ok := Actions[action]
	return ok
}
