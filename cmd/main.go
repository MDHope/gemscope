package main

import (
	"github.com/MDHope/gemscope/internal/browser"
	gemini_client "github.com/MDHope/gemscope/internal/gemini-client"
)

func main() {
	client, err := gemini_client.NewGeminiClient()
	if err != nil {
		panic(err)
	}
	browser.Start(client)
}
