package domain

// CategoryType represents the type of category
type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"  // For income transactions
	CategoryTypeExpense CategoryType = "expense" // For expense transactions
	CategoryTypeBoth    CategoryType = "both"    // Can be used for both income and expense
)

// ValidateCategoryType validates category type
func ValidateCategoryType(t CategoryType) bool {
	switch t {
	case CategoryTypeIncome, CategoryTypeExpense, CategoryTypeBoth:
		return true
	}
	return false
}

// DefaultCategories returns a list of system-provided default categories
func DefaultExpenseCategories() []struct {
	Name        string
	Description string
	Icon        string
	Color       string
	Type        CategoryType
} {
	return []struct {
		Name        string
		Description string
		Icon        string
		Color       string
		Type        CategoryType
	}{
		// Essential expenses
		{Name: "Food & Dining", Description: "Groceries, restaurants, cafes", Icon: "restaurant", Color: "#FF6B6B", Type: CategoryTypeExpense},
		{Name: "Transportation", Description: "Gas, public transport, parking", Icon: "directions_car", Color: "#4ECDC4", Type: CategoryTypeExpense},
		{Name: "Housing", Description: "Rent, mortgage, utilities", Icon: "home", Color: "#95E1D3", Type: CategoryTypeExpense},
		{Name: "Healthcare", Description: "Medical, dental, pharmacy", Icon: "local_hospital", Color: "#F38181", Type: CategoryTypeExpense},
		{Name: "Utilities", Description: "Electricity, water, internet", Icon: "bolt", Color: "#FFA726", Type: CategoryTypeExpense},

		// Lifestyle
		{Name: "Shopping", Description: "Clothing, electronics, personal items", Icon: "shopping_bag", Color: "#BA68C8", Type: CategoryTypeExpense},
		{Name: "Entertainment", Description: "Movies, concerts, hobbies", Icon: "movie", Color: "#FFB74D", Type: CategoryTypeExpense},
		{Name: "Travel", Description: "Flights, hotels, vacation", Icon: "flight", Color: "#81C784", Type: CategoryTypeExpense},
		{Name: "Education", Description: "Tuition, courses, books", Icon: "school", Color: "#64B5F6", Type: CategoryTypeExpense},
		{Name: "Personal Care", Description: "Haircut, spa, cosmetics", Icon: "face", Color: "#F06292", Type: CategoryTypeExpense},

		// Financial
		{Name: "Insurance", Description: "Health, life, auto insurance", Icon: "shield", Color: "#9575CD", Type: CategoryTypeExpense},
		{Name: "Debt Payment", Description: "Loan, credit card payments", Icon: "credit_card", Color: "#E57373", Type: CategoryTypeExpense},
		{Name: "Savings", Description: "Emergency fund, savings account", Icon: "savings", Color: "#4DB6AC", Type: CategoryTypeExpense},
		{Name: "Investment", Description: "Stocks, bonds, crypto", Icon: "trending_up", Color: "#7986CB", Type: CategoryTypeExpense},

		// Others
		{Name: "Gifts & Donations", Description: "Presents, charity", Icon: "card_giftcard", Color: "#FF8A65", Type: CategoryTypeExpense},
		{Name: "Pets", Description: "Pet food, vet, grooming", Icon: "pets", Color: "#A1887F", Type: CategoryTypeExpense},
		{Name: "Other", Description: "Miscellaneous expenses", Icon: "more_horiz", Color: "#90A4AE", Type: CategoryTypeExpense},
	}
}

// DefaultIncomeCategories returns default income categories
func DefaultIncomeCategories() []struct {
	Name        string
	Description string
	Icon        string
	Color       string
	Type        CategoryType
} {
	return []struct {
		Name        string
		Description string
		Icon        string
		Color       string
		Type        CategoryType
	}{
		{Name: "Salary", Description: "Monthly salary, wages", Icon: "attach_money", Color: "#66BB6A", Type: CategoryTypeIncome},
		{Name: "Business", Description: "Business income, self-employed", Icon: "business_center", Color: "#42A5F5", Type: CategoryTypeIncome},
		{Name: "Investment Returns", Description: "Dividends, interest, capital gains", Icon: "show_chart", Color: "#26A69A", Type: CategoryTypeIncome},
		{Name: "Freelance", Description: "Freelance work, gigs", Icon: "work", Color: "#AB47BC", Type: CategoryTypeIncome},
		{Name: "Rental Income", Description: "Property rental", Icon: "apartment", Color: "#78909C", Type: CategoryTypeIncome},
		{Name: "Bonus", Description: "Work bonus, commission", Icon: "star", Color: "#FFCA28", Type: CategoryTypeIncome},
		{Name: "Gift Received", Description: "Monetary gifts", Icon: "redeem", Color: "#FF7043", Type: CategoryTypeIncome},
		{Name: "Refund", Description: "Tax refund, cashback", Icon: "undo", Color: "#26C6DA", Type: CategoryTypeIncome},
		{Name: "Other Income", Description: "Miscellaneous income", Icon: "more_horiz", Color: "#9CCC65", Type: CategoryTypeIncome},
	}
}
