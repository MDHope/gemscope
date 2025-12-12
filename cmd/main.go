package main

import (
	"github.com/MDHope/gemscope/internal/browser"
	gemini_client "github.com/MDHope/gemscope/internal/gemini-client"
)

// "fmt"
// "os"
//

func main() {
	client, err := gemini_client.NewGeminiClient()
	if err != nil {
		panic(err)
	}
	browser.Start(client)

	// content, _ := os.ReadFile("test_gemtext.txt")
	// node := gemtext.Parse(string(content))
	// fmt.Println(node.Print())
	// if len(os.Args) < 2 {
	// 	fmt.Println("Usage: go run main.go gemini://example.com/path")
	// 	os.Exit(1)
	// }
	//
	// response, err := client.Fetch(os.Args[1])
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	// 	os.Exit(1)
	// }
	//
	// fmt.Printf("%d: %s\n", response.Status, response.Meta)
	// fmt.Print(response.Body)
}
