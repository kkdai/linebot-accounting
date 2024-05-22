package main

import "log"

func recordExpense(name string, date string, amount float64, category string) map[string]any {
	// This hypothetical API returns a JSON such as:
	// {"date":"2024-04-17","amount":50.0,"category":"Food","status":"Success"}

	expense := Expense{
		Name:     name,
		Category: category,
		Amount:   int(amount),
		Date:     date,
	}

	// Insert the expense to the database.
	if err := fireDB.InsertDB(expense); err != nil {
		log.Println("Storage save err:", err)
	}

	return map[string]any{
		"name":     name,
		"date":     date,
		"amount":   amount,
		"category": category,
		"status":   "Success",
	}
}
