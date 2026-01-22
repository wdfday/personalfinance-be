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

// DefaultCategoryDef represents a default category definition
type DefaultCategoryDef struct {
	Name          string
	Description   string
	Icon          string
	Color         string
	Type          CategoryType
	SubCategories []DefaultCategoryDef
}

// DefaultExpenseCategories returns a list of system-provided default expense categories (3-level hierarchy)
func DefaultExpenseCategories() []DefaultCategoryDef {
	return []DefaultCategoryDef{
		// 0. Buffer & Misc (Level 0) - dùng làm khoản đệm / chi phí linh hoạt tổng hợp
		{
			Name: "Buffer & Misc", Description: "General buffer and miscellaneous expenses", Icon: "savings", Color: "#9E9E9E", Type: CategoryTypeExpense,
			SubCategories: []DefaultCategoryDef{
				{
					Name: "Budget Buffer", Description: "Flexible buffer for unexpected or overflow expenses", Icon: "account_balance_wallet", Color: "#757575", Type: CategoryTypeExpense,
				},
			},
		},

		// 1. Home & Utilities (Level 0)
		{
			Name: "Home & Utilities", Description: "Housing and utility costs", Icon: "home", Color: "#4FC3F7", Type: CategoryTypeExpense,
			SubCategories: []DefaultCategoryDef{
				// Level 1: Housing
				{
					Name: "Housing", Description: "Housing costs", Icon: "account_balance", Color: "#0288D1", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Rent", Description: "Monthly rent payment", Icon: "key", Color: "#0277BD", Type: CategoryTypeExpense},
						{Name: "Mortgage", Description: "Mortgage payment", Icon: "home_work", Color: "#01579B", Type: CategoryTypeExpense},
						{Name: "Property Tax", Description: "Property tax", Icon: "gavel", Color: "#004D40", Type: CategoryTypeExpense},
						{Name: "Home Insurance", Description: "Home insurance premium", Icon: "verified_user", Color: "#006064", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Utilities
				{
					Name: "Utilities", Description: "Utility bills", Icon: "power", Color: "#FFD600", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Electricity", Description: "Electricity bill", Icon: "electric_bolt", Color: "#FBC02D", Type: CategoryTypeExpense},
						{Name: "Water", Description: "Water bill", Icon: "water_drop", Color: "#03A9F4", Type: CategoryTypeExpense},
						{Name: "Gas", Description: "Gas bill", Icon: "local_fire_department", Color: "#FF6F00", Type: CategoryTypeExpense},
						{Name: "Internet & TV", Description: "Internet and cable TV", Icon: "wifi", Color: "#7E57C2", Type: CategoryTypeExpense},
						{Name: "Phone", Description: "Mobile and landline", Icon: "phone", Color: "#5E35B1", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Maintenance
				{
					Name: "Maintenance", Description: "Home maintenance", Icon: "build", Color: "#607D8B", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Repairs", Description: "Home repairs", Icon: "handyman", Color: "#546E7A", Type: CategoryTypeExpense},
						{Name: "Cleaning", Description: "Cleaning services", Icon: "cleaning_services", Color: "#78909C", Type: CategoryTypeExpense},
						{Name: "Gardening", Description: "Garden and lawn care", Icon: "yard", Color: "#66BB6A", Type: CategoryTypeExpense},
					},
				},
			},
		},
		// 2. Food & Dining (Level 0)
		{
			Name: "Food & Dining", Description: "Food and eating out", Icon: "restaurant", Color: "#FF6B6B", Type: CategoryTypeExpense,
			SubCategories: []DefaultCategoryDef{
				// Level 1: Groceries
				{
					Name: "Groceries", Description: "Grocery shopping", Icon: "shopping_cart", Color: "#EF5350", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Supermarket", Description: "Supermarket shopping", Icon: "store", Color: "#E53935", Type: CategoryTypeExpense},
						{Name: "Fresh Market", Description: "Traditional markets", Icon: "storefront", Color: "#C62828", Type: CategoryTypeExpense},
						{Name: "Convenience Store", Description: "Convenience store", Icon: "local_convenience_store", Color: "#B71C1C", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Dining Out
				{
					Name: "Dining Out", Description: "Restaurant and eating out", Icon: "restaurant_menu", Color: "#EC407A", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Restaurants", Description: "Sit-down restaurants", Icon: "dinner_dining", Color: "#D81B60", Type: CategoryTypeExpense},
						{Name: "Fast Food", Description: "Fast food and quick service", Icon: "fastfood", Color: "#C2185B", Type: CategoryTypeExpense},
						{Name: "Delivery", Description: "Food delivery", Icon: "delivery_dining", Color: "#AD1457", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Beverages
				{
					Name: "Beverages", Description: "Coffee, tea, drinks", Icon: "local_cafe", Color: "#F48FB1", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Coffee & Tea", Description: "Coffee shops and tea houses", Icon: "coffee", Color: "#F06292", Type: CategoryTypeExpense},
						{Name: "Bars & Pubs", Description: "Bars and pubs", Icon: "local_bar", Color: "#E91E63", Type: CategoryTypeExpense},
					},
				},
			},
		},
		// 3. Transportation (Level 0)
		{
			Name: "Transportation", Description: "Moving around", Icon: "directions_car", Color: "#FFA726", Type: CategoryTypeExpense,
			SubCategories: []DefaultCategoryDef{
				// Level 1: Vehicle
				{
					Name: "Vehicle", Description: "Personal vehicle costs", Icon: "directions_car", Color: "#FB8C00", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Gas/Fuel", Description: "Fuel for vehicle", Icon: "local_gas_station", Color: "#E65100", Type: CategoryTypeExpense},
						{Name: "Parking", Description: "Parking fees", Icon: "local_parking", Color: "#EF6C00", Type: CategoryTypeExpense},
						{Name: "Tolls", Description: "Highway tolls", Icon: "toll", Color: "#E64A19", Type: CategoryTypeExpense},
						{Name: "Car Maintenance", Description: "Vehicle maintenance and repair", Icon: "car_repair", Color: "#D84315", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Public Transport
				{
					Name: "Public Transport", Description: "Public transportation", Icon: "directions_bus", Color: "#FFA000", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Bus", Description: "Bus fares", Icon: "directions_bus", Color: "#FF8F00", Type: CategoryTypeExpense},
						{Name: "Train/Metro", Description: "Train and metro", Icon: "train", Color: "#FF6F00", Type: CategoryTypeExpense},
						{Name: "Taxi/Ride Share", Description: "Grab, Uber, Taxi", Icon: "local_taxi", Color: "#FFB74D", Type: CategoryTypeExpense},
					},
				},
			},
		},
		// 4. Shopping (Level 0)
		{
			Name: "Shopping", Description: "Buying things", Icon: "shopping_bag", Color: "#AB47BC", Type: CategoryTypeExpense,
			SubCategories: []DefaultCategoryDef{
				// Level 1: Personal Items
				{
					Name: "Personal Items", Description: "Personal shopping", Icon: "shopping_bag", Color: "#9C27B0", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Clothing", Description: "Clothes and accessories", Icon: "checkroom", Color: "#8E24AA", Type: CategoryTypeExpense},
						{Name: "Shoes", Description: "Footwear", Icon: "directions_walk", Color: "#7B1FA2", Type: CategoryTypeExpense},
						{Name: "Accessories", Description: "Bags, jewelry", Icon: "watch", Color: "#6A1B9A", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Electronics
				{
					Name: "Electronics", Description: "Tech and gadgets", Icon: "devices", Color: "#5E35B1", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Computers", Description: "Laptops, desktops", Icon: "computer", Color: "#512DA8", Type: CategoryTypeExpense},
						{Name: "Mobile Devices", Description: "Phones, tablets", Icon: "smartphone", Color: "#4527A0", Type: CategoryTypeExpense},
						{Name: "Home Appliances", Description: "Appliances for home", Icon: "kitchen", Color: "#311B92", Type: CategoryTypeExpense},
					},
				},
			},
		},
		// 5. Health & Wellness (Level 0)
		{
			Name: "Health & Wellness", Description: "Health and personal care", Icon: "favorite", Color: "#26A69A", Type: CategoryTypeExpense,
			SubCategories: []DefaultCategoryDef{
				// Level 1: Medical
				{
					Name: "Medical", Description: "Healthcare services", Icon: "medical_services", Color: "#00897B", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Doctor Visits", Description: "Doctor consultations", Icon: "local_hospital", Color: "#00796B", Type: CategoryTypeExpense},
						{Name: "Medicine", Description: "Medicines and supplements", Icon: "medication", Color: "#00695C", Type: CategoryTypeExpense},
						{Name: "Dental", Description: "Dental care", Icon: "medication_liquid", Color: "#004D40", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Fitness
				{
					Name: "Fitness", Description: "Sports and fitness", Icon: "fitness_center", Color: "#4DB6AC", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Gym Membership", Description: "Gym and fitness center", Icon: "sports_gymnastics", Color: "#26A69A", Type: CategoryTypeExpense},
						{Name: "Sports Equipment", Description: "Sports gear", Icon: "sports_soccer", Color: "#009688", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Personal Care
				{
					Name: "Personal Care", Description: "Beauty and grooming", Icon: "face", Color: "#80CBC4", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Haircut & Salon", Description: "Hair and beauty salon", Icon: "content_cut", Color: "#4DB6AC", Type: CategoryTypeExpense},
						{Name: "Cosmetics", Description: "Makeup and skincare", Icon: "spa", Color: "#26A69A", Type: CategoryTypeExpense},
					},
				},
			},
		},
		// 6. Entertainment & Travel (Level 0)
		{
			Name: "Entertainment & Travel", Description: "Fun and trips", Icon: "flight_takeoff", Color: "#4DB6AC", Type: CategoryTypeExpense,
			SubCategories: []DefaultCategoryDef{
				// Level 1: Entertainment
				{
					Name: "Entertainment", Description: "Fun activities", Icon: "theaters", Color: "#009688", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Movies & Events", Description: "Cinema, concerts", Icon: "movie", Color: "#00897B", Type: CategoryTypeExpense},
						{Name: "Games", Description: "Video games, board games", Icon: "sports_esports", Color: "#00796B", Type: CategoryTypeExpense},
						{Name: "Hobbies", Description: "Hobby supplies", Icon: "palette", Color: "#00695C", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Travel
				{
					Name: "Travel", Description: "Trips and vacations", Icon: "flight", Color: "#26A69A", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Flights", Description: "Air travel", Icon: "flight_takeoff", Color: "#26A69A", Type: CategoryTypeExpense},
						{Name: "Hotels", Description: "Accommodation", Icon: "hotel", Color: "#009688", Type: CategoryTypeExpense},
						{Name: "Tours & Activities", Description: "Tourist activities", Icon: "tour", Color: "#00897B", Type: CategoryTypeExpense},
					},
				},
			},
		},
		// 7. Education & Self-Dev (Level 0)
		{
			Name: "Education & Self-Dev", Description: "Learning and growth", Icon: "school", Color: "#5C6BC0", Type: CategoryTypeExpense,
			SubCategories: []DefaultCategoryDef{
				// Level 1: Formal Education
				{
					Name: "Formal Education", Description: "School and university", Icon: "school", Color: "#3949AB", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Tuition", Description: "School fees", Icon: "account_balance", Color: "#303F9F", Type: CategoryTypeExpense},
						{Name: "Books & Materials", Description: "Textbooks and supplies", Icon: "menu_book", Color: "#283593", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Self Development
				{
					Name: "Self Development", Description: "Personal learning", Icon: "laptop_chromebook", Color: "#7986CB", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Online Courses", Description: "Online learning platforms", Icon: "video_library", Color: "#5C6BC0", Type: CategoryTypeExpense},
						{Name: "Workshops", Description: "Workshops and seminars", Icon: "groups", Color: "#3F51B5", Type: CategoryTypeExpense},
					},
				},
			},
		},
		// 8. Financial & Obligations (Level 0)
		{
			Name: "Financial & Obligations", Description: "Financial duties", Icon: "account_balance_wallet", Color: "#78909C", Type: CategoryTypeExpense,
			SubCategories: []DefaultCategoryDef{
				// Level 1: Insurance
				{
					Name: "Insurance", Description: "Insurance premiums", Icon: "shield", Color: "#546E7A", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Life Insurance", Description: "Life insurance premium", Icon: "favorite", Color: "#455A64", Type: CategoryTypeExpense},
						{Name: "Health Insurance", Description: "Health insurance premium", Icon: "local_hospital", Color: "#37474F", Type: CategoryTypeExpense},
						{Name: "Vehicle Insurance", Description: "Car/bike insurance", Icon: "directions_car", Color: "#263238", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Taxes & Debts
				{
					Name: "Taxes & Debts", Description: "Taxes and loan payments", Icon: "receipt_long", Color: "#607D8B", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Income Tax", Description: "Income tax payment", Icon: "gavel", Color: "#546E7A", Type: CategoryTypeExpense},
						{Name: "Debt Payment", Description: "Loan repayments", Icon: "credit_card", Color: "#455A64", Type: CategoryTypeExpense},
					},
				},
				// Level 1: Giving
				{
					Name: "Giving", Description: "Gifts and donations", Icon: "volunteer_activism", Color: "#90A4AE", Type: CategoryTypeExpense,
					SubCategories: []DefaultCategoryDef{
						{Name: "Gifts", Description: "Gifts for others", Icon: "redeem", Color: "#78909C", Type: CategoryTypeExpense},
						{Name: "Donations", Description: "Charitable donations", Icon: "volunteer_activism", Color: "#607D8B", Type: CategoryTypeExpense},
					},
				},
			},
		},
	}
}

// DefaultIncomeCategories returns default income categories (3-level hierarchy)
func DefaultIncomeCategories() []DefaultCategoryDef {
	return []DefaultCategoryDef{
		// 1. Active Income (Level 0)
		{
			Name: "Active Income", Description: "Money earned from working", Icon: "work", Color: "#66BB6A", Type: CategoryTypeIncome,
			SubCategories: []DefaultCategoryDef{
				// Level 1: Employment
				{
					Name: "Employment", Description: "Regular employment income", Icon: "business_center", Color: "#43A047", Type: CategoryTypeIncome,
					SubCategories: []DefaultCategoryDef{
						{Name: "Salary", Description: "Monthly salary", Icon: "payments", Color: "#388E3C", Type: CategoryTypeIncome},
						{Name: "Bonus", Description: "Performance bonus", Icon: "star", Color: "#2E7D32", Type: CategoryTypeIncome},
						{Name: "Overtime", Description: "Overtime pay", Icon: "schedule", Color: "#1B5E20", Type: CategoryTypeIncome},
					},
				},
				// Level 1: Self Employment
				{
					Name: "Self Employment", Description: "Freelance and contract work", Icon: "design_services", Color: "#66BB6A", Type: CategoryTypeIncome,
					SubCategories: []DefaultCategoryDef{
						{Name: "Freelance", Description: "Freelance projects", Icon: "laptop", Color: "#4CAF50", Type: CategoryTypeIncome},
						{Name: "Consulting", Description: "Consulting fees", Icon: "support_agent", Color: "#43A047", Type: CategoryTypeIncome},
					},
				},
			},
		},
		// 2. Investment & Business (Level 0)
		{
			Name: "Investment & Business", Description: "Passive income sources", Icon: "trending_up", Color: "#29B6F6", Type: CategoryTypeIncome,
			SubCategories: []DefaultCategoryDef{
				// Level 1: Investment Income
				{
					Name: "Investment Income", Description: "Returns from investments", Icon: "savings", Color: "#039BE5", Type: CategoryTypeIncome,
					SubCategories: []DefaultCategoryDef{
						{Name: "Interest", Description: "Bank interest", Icon: "account_balance", Color: "#0288D1", Type: CategoryTypeIncome},
						{Name: "Dividends", Description: "Stock dividends", Icon: "pie_chart", Color: "#0277BD", Type: CategoryTypeIncome},
						{Name: "Capital Gains", Description: "Investment profits", Icon: "show_chart", Color: "#01579B", Type: CategoryTypeIncome},
					},
				},
				// Level 1: Business Income
				{
					Name: "Business Income", Description: "Business and rental", Icon: "store", Color: "#4FC3F7", Type: CategoryTypeIncome,
					SubCategories: []DefaultCategoryDef{
						{Name: "Business Profit", Description: "Profit from business", Icon: "storefront", Color: "#29B6F6", Type: CategoryTypeIncome},
						{Name: "Rental Income", Description: "Property rental income", Icon: "house", Color: "#03A9F4", Type: CategoryTypeIncome},
					},
				},
			},
		},
		// 3. Other Income (Level 0)
		{
			Name: "Other Income", Description: "Miscellaneous income", Icon: "category", Color: "#BA68C8", Type: CategoryTypeIncome,
			SubCategories: []DefaultCategoryDef{
				// Level 1: One-time Income
				{
					Name: "One-time Income", Description: "Non-recurring income", Icon: "monetization_on", Color: "#AB47BC", Type: CategoryTypeIncome,
					SubCategories: []DefaultCategoryDef{
						{Name: "Gifts", Description: "Monetary gifts received", Icon: "redeem", Color: "#9C27B0", Type: CategoryTypeIncome},
						{Name: "Tax Refunds", Description: "Tax refund", Icon: "replay", Color: "#8E24AA", Type: CategoryTypeIncome},
						{Name: "Selling Items", Description: "Selling used items", Icon: "sell", Color: "#7B1FA2", Type: CategoryTypeIncome},
					},
				},
			},
		},
	}
}
