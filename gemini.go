package main


// Today I learned: had used {} instead of () for import statement
import (
	"context"
	"fmt"
	"os"
	"log"

	"google.golang.org/genai" // TODO [answer] is this a named import?
)

type GeminiChat struct {
	client *genai.Client
	chat *genai.Chat
}

// TODO (remove this)------------------------  has a very return tuple vibe from Python
func NewGeminiChatSession(ctx context.Context) (*GeminiChat, error) {
	
	// The variables declared in .env will become available here thanks 
	// to the godotenv library

	apiKey := os.Getenv("GOOGLE_API_KEY") 
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY not set")
	}
	
	// client, err := genai.NewClient(ctx, genai.WithAPIKey(apiKey))
	// client, err := genai.NewClient(ctx, &genai.ClientConfig{Backend: genai.BackendGeminiAPI, HTTPOptions: genai.HTTPOptions{APIVersion: "v1beta"}})
	
	client, err := genai.NewClient(
		ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		return nil, err
	}
	
	// model := client.GenerativeModel("gemini-2.0-flash-lite")
	// chat := model.StartChat()
	chat, err := client.Chats.Create(ctx, "gemini-2.0-flash", nil, nil)
	if err != nil {
		log.Fatal("Error while creating Gemini chat", err)
		return nil, err
	}
	return &GeminiChat{
		client: client,
		chat: chat,
	}, nil
}


// TODO [understand] this has a very Kotlin extension function vibe
func (g *GeminiChat) SendMessage(ctx context.Context, msg string) (string, error) {
	// resp is a GenerateContentResponse
	resp, err := g.chat.SendMessage(ctx, genai.Part{Text: msg})
	if err != nil {
		return "", err
	}

	// Return first candidate
	if len(resp.Candidates) > 0 {
		// TODO: figure out what this mumbo jumbo means
		// TODO: extract the string to return in a temporary variable
		var response string
		log.Println("initial response for websocket is", response)
		for _, candidate := range resp.Candidates {
			// candidate is a ptr to a Candidate
			for _, part := range candidate.Content.Parts {
				// part is a ptr to a part
				if part.Text != "" {
					response += part.Text
				}
			}
		}
		return response, nil
		// return fmt.Sprintf("%q", resp.Candidates[0].Content.Parts[0]), nil
	}	

	return "No response", nil
}