package main

func recordExpense(date string, amount float64, category string) map[string]any {
	// This hypothetical API returns a JSON such as:
	// {"date":"2024-04-17","amount":50.0,"category":"Food","status":"Success"}
	return map[string]any{
		"date":     date,
		"amount":   amount,
		"category": category,
		"status":   "Success",
	}
}
