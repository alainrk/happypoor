package ai

import (
	"bytes"
	"text/template"
)

// LLM Template for Expenses
const LLMExpensePromptTemplate = `You are a financial transaction parser. Your task is to analyze the input text and extract the following information:
- The category of the transaction
- The amount spent or received
- A brief description of the transaction

Format the result as a JSON object with the following structure:
{ "category": "Category", "amount": 12.34, "description": "Description" }

Available categories (use ONLY these):
"Car", "Clothes", "Grocery", "House", "Bills", "Entertainment", "Sport", "EatingOut", "Transport", "Learning", "Toiletry", "Health", "Tech", "Gifts", "Travel", "Pets", "OtherExpenses"

Follow these rules:
1. For category selection:
   - First try to find the category directly mentioned in the text (accounting for typos/synonyms)
   - If no category is directly mentioned, infer it from the description
   - If category cannot be determined, use "OtherExpenses"
2. For description:
   - Use the main item mentioned in the text
   - Capitalize the first letter of the description
   - If no item is mentioned, use text of the category
3. For amount:
   - Convert any amount to standard decimal notation with a period (not comma) as decimal separator
   - Return as a number (not a string) with at most 2 decimal places
   - If no amount is mentioned, use 0

Examples:
- "bread 5 euro an 20, grocery" → { "category": "Grocery", "amount": 5.2, "description": "Bread" }
- "pam 4.31 grocertw" → { "category": "Grocery", "amount": 4.31, "description": "Pam" }
- "car 25,30" → { "category": "Car", "amount": 25.3, "description": "Car" }
- "34 usd 23-04" → { "category": "OtherExpenses", "amount": 34, "description": "OtherExpenses" }
- "Great sea food 12 euro e 25" → { "category": "EatingOut", "amount": 12.25, "description": "Great see food" }

IMPORTANT: Respond with ONLY the JSON object but without markdown syntax. Your answer is plaintext being JSON to be parsed as it is, don't include the triple backticks syntax or anything similar.

User input:
{{.UserText}}
`

// LLM Template for Incomes
const LLMIncomePromptTemplate = `You are a financial transaction parser. Your task is to analyze the input text and extract the following information:
- The category of the transaction
- The amount spent or received
- A brief description of the transaction

Format the result as a JSON object with the following structure:
{ "category": "Category", "amount": 12.34, "description": "Description" }

Available categories (use ONLY these):
"Salary", "OtherIncomes"

Follow these rules:
1. For category selection:
   - First try to find the category directly mentioned in the text (accounting for typos/synonyms)
   - If no category is directly mentioned, infer it from the description and prefer "Salary" only when the user use it or with a synonym in any language
   - If category cannot be determined, use "OtherIncomes", for example for "ticket restaurants", "refund amazon", etc.
2. For description:
   - Use the main item mentioned in the text
   - Capitalize the first letter of the description
   - If no item is mentioned, use text of the category
3. For amount:
   - Convert any amount to standard decimal notation with a period (not comma) as decimal separator
   - Return as a number (not a string) with at most 2 decimal places
   - If no amount is mentioned, use 0

Examples:
- "250k earned from job" → { "category": "Salary", "amount": 250000, "description": "From job" }
- "salayr 340 and 34 august" → { "category": "Salary", "amount": 340.34, "description": "August" }
- "ticket reastants 245 dollars" → { "category": "OtherIncomes", "amount": 245, "description": "Ticket restaurants" }
- "gained income 231 and 32 euro 03-04" → { "category": "Salary", "amount": 231.32, "description": "Salary" }

IMPORTANT: Respond with ONLY the JSON object but without markdown syntax. Your answer is plaintext being JSON to be parsed as it is, don't include the triple backticks syntax or anything similar.

User input:
{{.UserText}}
`

// GeneratePrompt creates the complete prompt by filling in the template with user input
func GeneratePrompt(userText string, promptTemplate string) (string, error) {
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", err
	}

	data := struct {
		UserText string
	}{
		UserText: userText,
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
