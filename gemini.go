package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiApp struct {
	geminiKey string
	ctx       context.Context
	client    *genai.Client
}

var expenseTrackingTool *genai.Tool
var expenseListingTool *genai.Tool

// Init Gemini API
func InitGemini(key string) *GeminiApp {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		log.Fatal(err)
	}

	expenseTrackingTool = &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "recordExpense",
			Description: "Record an expense with date, amount, and category",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"name": {
						Type:        genai.TypeString,
						Description: "The name of the expense",
					},
					"date": {
						Type:        genai.TypeString,
						Description: "The date of the expense in YYYY-MM-DD format",
					},
					"amount": {
						Type:        genai.TypeNumber,
						Description: "The amount of the expense",
					},
					"category": {
						Type:        genai.TypeString,
						Description: "The category of the expense, it could be one of following (食, 衣, 住, 行)",
					},
				},
				Required: []string{"name,", "date", "amount", "category"},
			},
		}},
	}

	expenseListingTool = &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "listAllExpense",
			Description: "List all expenses within a specific date range, or all expenses if no dates are specified",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"startDate": {
						Type:        genai.TypeString,
						Description: "Start date of the period in YYYY-MM-DD format (optional)",
					},
					"endDate": {
						Type:        genai.TypeString,
						Description: "End date of the period in YYYY-MM-DD format (optional)",
					},
				},
			},
		}},
	}
	return &GeminiApp{key, ctx, client}
}

// Gemini Text: Input a prompt and get the response string.
func (app *GeminiApp) GeminiImage(imgData []byte, prompt string) (string, error) {
	model := app.client.GenerativeModel("gemini-pro-vision")
	// Set the temperature to 0.8 for a balance between creativity and coherence.
	value := float32(0.8)
	model.Temperature = &value
	data := []genai.Part{
		genai.ImageData("png", imgData),
		genai.Text(prompt),
	}
	log.Println("Begin processing image...")
	resp, err := model.GenerateContent(app.ctx, data...)
	log.Println("Finished processing image...", resp)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	return printResponse(resp), nil
}

// Gemini Chat Complete: Iput a prompt and get the response string.
func (app *GeminiApp) GeminiChatComplete(req string) string {
	model := app.client.GenerativeModel("gemini-1.5-flash-latest")
	value := float32(0.8)
	model.Temperature = &value
	cs := model.StartChat()

	send := func(msg string) *genai.GenerateContentResponse {
		fmt.Printf("== Me: %s\n== Model:\n", msg)
		res, err := cs.SendMessage(app.ctx, genai.Text(msg))
		if err != nil {
			log.Fatal(err)
		}
		return res
	}

	res := send(req)
	return printResponse(res)
}

// Gemini Function Call: Input a prompt and get the response string.
func (app *GeminiApp) GeminiFunctionCall(prompt string) string {
	// Add timestamp for this prompt.
	timelocal, _ := time.LoadLocation("Asia/Taipei")
	time.Local = timelocal
	curNow := time.Now().Local().String()
	prompt = prompt + " 本地時間: " + curNow

	// Use a model that supports function calling, like Gemini 1.0 Pro.
	model := app.client.GenerativeModel("gemini-1.5-flash-latest")

	// Specify the function declaration.
	model.Tools = []*genai.Tool{expenseTrackingTool, expenseListingTool}

	// Start new chat session.
	session := model.StartChat()

	// prompt = "Can you record an expense of 50 dollars for workout out on 2024-05-17?"

	// Send the message to the generative model.
	resp, err := session.SendMessage(app.ctx, genai.Text(prompt))
	if err != nil {
		log.Fatalf("Error sending message: %v\n", err)
	}

	// Check that you got the expected function call back.
	part := resp.Candidates[0].Content.Parts[0]
	funcall, ok := part.(genai.FunctionCall)
	if !ok {
		log.Fatalf("Expected type FunctionCall, got %T", part)
	}
	if g, e := funcall.Name, expenseTrackingTool.FunctionDeclarations[0].Name; g != e {
		log.Fatalf("Expected FunctionCall.Name %q, got %q", e, g)
	}
	fmt.Printf("Received function call response:\n %s \n %s \n", part.(genai.FunctionCall).Name, part.(genai.FunctionCall).Args)

	//Accordig to funccall Name and Args we can call the function
	switch part.(genai.FunctionCall).Name {
	case "recordExpense":
		log.Println("Calling recordExpense function...")
		args := part.(genai.FunctionCall).Args
		name := args["name"]
		date := args["date"]
		amount := args["amount"]
		category := args["category"]

		log.Println("date: ", date, "amount: ", amount, "category: ", category)

		// Call the hypothetical API to record the expense.
		apiResult := recordExpense(name.(string), date.(string), amount.(float64), category.(string))
		// Send the hypothetical API result back to the generative model.
		fmt.Printf("Sending API result:\n%q\n\n", apiResult)
		resp, err = session.SendMessage(app.ctx, genai.FunctionResponse{
			Name:     expenseTrackingTool.FunctionDeclarations[0].Name,
			Response: apiResult,
		})
		if err != nil {
			log.Fatalf("Error sending message: %v\n", err)
		}

		// Show the model's response, which is expected to be text.
		return printResponse(resp)
	case "listAllExpense":
		log.Println("Calling recordExpense function...")
		args := part.(genai.FunctionCall).Args
		startDate := args["startDate"]
		endDate := args["endDate"]
		log.Println("startDate: ", startDate, " endDate: ", endDate)

		// Call the hypothetical API to list all the expense.
		apiResult := listAllExpense(startDate.(string), endDate.(string))

		// Send the hypothetical API result back to the generative model.
		fmt.Printf("Sending API result:\n%q\n\n", apiResult)
		resp, err = session.SendMessage(app.ctx, genai.FunctionResponse{
			Name:     expenseListingTool.FunctionDeclarations[0].Name,
			Response: apiResult,
		})
		if err != nil {
			log.Fatalf("Error sending message: %v\n", err)
		}

	}

	// If no function call was made, return the response as text.
	return app.GeminiChatComplete(prompt)
}

// removeFirstAndLastLine takes a string and removes the first and last lines.
func removeFirstAndLastLine(s string) string {
	// Split the string into lines.
	lines := strings.Split(s, "\n")

	// If there are less than 3 lines, return an empty string because removing the first and last would leave nothing.
	if len(lines) < 3 {
		return ""
	}

	// Join the lines back together, skipping the first and last lines.
	return strings.Join(lines[1:len(lines)-1], "\n")
}

// Print the response
func printResponse(resp *genai.GenerateContentResponse) string {
	var ret string
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			ret = ret + fmt.Sprintf("%v", part)
			fmt.Println(part)
		}
	}
	return ret
}
