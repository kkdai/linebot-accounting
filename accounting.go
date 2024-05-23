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

// listAllExpense: List all expenses within the specified date range.
func listAllExpense(startDate string, endDate string) map[string]any {
	filteredExpenses := make(map[string]any)
	// Get all expenses from the database.
	var expenses map[string]Expense
	if err := fireDB.GetFromDB(&expenses); err != nil {
		log.Println("Storage get err:", err)
		return nil
	}

	// Filter the expenses based on the specified date range.
	for _, expense := range expenses {
		expenseDate := expense.Date
		// Include the expense if it falls within the specified date range or if no dates are specified
		if (startDate == "" && endDate == "") || (expenseDate >= startDate && expenseDate <= endDate) {
			filteredExpenses[expense.Name+"-"+expense.Date] = expense
		}
	}
	return filteredExpenses
}
