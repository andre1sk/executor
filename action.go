package main

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
)

type Action interface {
	Run(in string) (out string, err error)
	Init() error
	Name() string
}

type ActionRunner interface {
	RunAction(action string, in string) (out string, err error)
	ActionExists(action string) bool
}

type SimpleActionRunner struct {
	actions map[string]Action
	logger  *log.Logger
}

// Verify interface compliance
var _ ActionRunner = (*SimpleActionRunner)(nil)

func NewSimpleActionRunner(logger *log.Logger) (*SimpleActionRunner, error) {
	actions := map[string]Action{
		"email_get_body":       &EmailBodyGetAction{},
		"str_extract_url":      &StrExtractURLAction{},
		"url_check_reputation": &URLCheckReputationAction{},
		"dummy":                &DummyAction{},
	}

	for _, action := range actions {
		if err := action.Init(); err != nil {
			return nil, fmt.Errorf("failed to init action %s: %v", action.Name(), err)
		}
	}
	return &SimpleActionRunner{
		logger:  logger,
		actions: actions,
	}, nil
}

func (a *SimpleActionRunner) RunAction(action string, in string) (out string, err error) {
	if act, ok := a.actions[action]; ok {
		return act.Run(in)
	}
	return "", fmt.Errorf("action %s not found", action)
}

func (a *SimpleActionRunner) ActionExists(action string) bool {
	_, ok := a.actions[action]
	return ok
}

type Email struct {
	EmailID string `json:"email-id"`
	Body    string `json:"body"`
}

type EmailBodyGetAction struct {
	emailBods map[string]string
}

// Verify interface compliance
var _ Action = (*EmailBodyGetAction)(nil)

func (e *EmailBodyGetAction) Run(in string) (out string, err error) {
	if bod, ok := e.emailBods[in]; ok {
		out = bod
		return
	}
	return "", fmt.Errorf("email body not found")
}

func (e *EmailBodyGetAction) Init() error {
	e.emailBods = make(map[string]string)
	var emails []Email

	err := ParseJsonFileInto("config/emails.json", &emails)
	if err != nil {
		return fmt.Errorf("failed to load emails: %v", err)
	}

	for _, email := range emails {
		e.emailBods[email.EmailID] = email.Body
	}

	return nil
}

func (e *EmailBodyGetAction) Name() string {
	return "email_get_body"
}

type StrExtractURLAction struct {
	findURL *regexp.Regexp
	hasHTTP *regexp.Regexp
}

// Verify interface compliance
var _ Action = (*StrExtractURLAction)(nil)

func (s *StrExtractURLAction) Run(in string) (out string, err error) {
	matches := s.findURL.FindAllString(in, -1)

	for _, match := range matches {
		if !s.hasHTTP.MatchString(match) {
			match = "http://" + match
		}

		parsedURL, err := url.Parse(match)
		if err == nil {
			return parsedURL.String(), nil
		}

	}

	return "", nil
}

func (s *StrExtractURLAction) Init() error {
	var err error
	s.findURL, err = regexp.Compile(`(?:https?://|www\.)[^\s]+`)
	if err != nil {
		return fmt.Errorf("failed to compile regex: %v", err)
	}
	s.hasHTTP, err = regexp.Compile(`^https?://`)
	if err != nil {
		return fmt.Errorf("failed to compile regex: %v", err)
	}
	return nil
}

func (s *StrExtractURLAction) Name() string {
	return "str_extract_url"
}

type URLCheckReputationAction struct{}

// Verify interface compliance
var _ Action = (*URLCheckReputationAction)(nil)

func (u *URLCheckReputationAction) Run(in string) (out string, err error) {
	out = "Good"
	return
}

func (u *URLCheckReputationAction) Init() error {
	return nil
}

func (u *URLCheckReputationAction) Name() string {
	return "url_check_reputation"
}

type DummyAction struct{}

// Verify interface compliance
var _ Action = (*DummyAction)(nil)

func (d *DummyAction) Run(in string) (out string, err error) {
	out = "Dummy"
	return
}

func (d *DummyAction) Init() error {
	return nil
}

func (d *DummyAction) Name() string {
	return "dummy"
}
