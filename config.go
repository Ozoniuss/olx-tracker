package main

//go:generate go tool genconfig -struct=Config -project=OlxTracker
type Config struct {
	OlxApiAuth OlxApiAuth
}

type OlxApiAuth struct {
	ClientID     string
	ClientSecret string
}
