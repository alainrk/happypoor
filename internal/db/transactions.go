package db

import (
	"cashout/internal/model"
	"fmt"
	"time"
)

// CreateTransaction creates a new transaction record
func (db *DB) CreateTransaction(transaction *model.Transaction) error {
	return db.conn.Create(transaction).Error
}

// GetTransactionByID retrieves an transaction by its ID
func (db *DB) GetTransactionByID(id int64) (*model.Transaction, error) {
	var transaction model.Transaction
	result := db.conn.Where("id = ?", id).First(&transaction)
	if result.Error != nil {
		return nil, result.Error
	}
	return &transaction, nil
}

// UpdateTransaction updates an existing transaction
func (db *DB) UpdateTransaction(transaction *model.Transaction) error {
	return db.conn.Save(transaction).Error
}

// DeleteTransaction deletes an transaction by ID (kept for backward compatibility)
func (db *DB) DeleteTransaction(id int64) error {
	return db.conn.Delete(&model.Transaction{}, id).Error
}

// DeleteTransactionByID deletes a transaction by its ID and optionally checks if it belongs to the given user
// If tgID is 0, the ownership check is skipped
func (db *DB) DeleteTransactionByID(id int64, tgID int64) error {
	query := db.conn.Model(&model.Transaction{})

	if tgID != 0 {
		// Add ownership check if tgID is provided
		query = query.Where("id = ? AND tg_id = ?", id, tgID)
	} else {
		query = query.Where("id = ?", id)
	}

	result := query.Delete(&model.Transaction{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("transaction not found or doesn't belong to user")
	}

	return nil
}

// GetUserTransactions retrieves all transactions for a user
func (db *DB) GetUserTransactions(tgID int64) ([]model.Transaction, error) {
	var transactions []model.Transaction
	result := db.conn.Where("tg_id = ?", tgID).Order("date DESC").Find(&transactions)
	if result.Error != nil {
		return nil, result.Error
	}
	return transactions, nil
}

// GetUserTransactionsByDateRange retrieves transactions for a user within a date range
func (db *DB) GetUserTransactionsByDateRange(tgID int64, startDate, endDate time.Time) ([]model.Transaction, error) {
	var transactions []model.Transaction
	result := db.conn.Where("tg_id = ? AND date BETWEEN ? AND ?",
		tgID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Order("date DESC").
		Find(&transactions)

	if result.Error != nil {
		return nil, result.Error
	}
	return transactions, nil
}

// GetUserTransactionsByDateRangePaginated retrieves paginated transactions for a user between two dates
func (db *DB) GetUserTransactionsByDateRangePaginated(tgID int64, startDate, endDate time.Time, offset, limit int) ([]model.Transaction, int64, error) {
	var transactions []model.Transaction
	var total int64

	// Get total count
	err := db.conn.Model(&model.Transaction{}).
		Where("tg_id = ? AND date BETWEEN ? AND ?",
			tgID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	result := db.conn.Where("tg_id = ? AND date BETWEEN ? AND ?",
		tgID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Order("date DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return transactions, total, nil
}

// GetUserTransactionsByMonth retrieves transactions for a user for a specific year and month
func (db *DB) GetUserTransactionsByMonth(tgID int64, year int, month int) ([]model.Transaction, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1) // Last day of the month

	return db.GetUserTransactionsByDateRange(tgID, startDate, endDate)
}

// GetUserTransactionsByCategory retrieves transactions for a user grouped by category
func (db *DB) GetUserTransactionsByCategory(tgID int64, startDate, endDate time.Time, transactionType model.TransactionType) (map[model.TransactionCategory]float64, error) {
	var results []struct {
		Category model.TransactionCategory
		Total    float64
	}

	query := db.conn.Table("transactions").
		Select("category, SUM(amount) as total").
		Where("tg_id = ? AND date BETWEEN ? AND ? AND type = ?",
			tgID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), transactionType).
		Group("category").
		Order("total DESC")

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	// Convert to map
	categoryTotals := make(map[model.TransactionCategory]float64)
	for _, result := range results {
		categoryTotals[result.Category] = result.Total
	}

	return categoryTotals, nil
}

// GetUserBalance calculates the total balance (income - transactions) for a user
func (db *DB) GetUserBalance(tgID int64, startDate, endDate time.Time) (float64, error) {
	var income float64
	var transaction float64

	// Get total income
	incomeQuery := db.conn.Table("transactions").
		Select("COALESCE(SUM(amount), 0) as total").
		Where("tg_id = ? AND date BETWEEN ? AND ? AND type = ?",
			tgID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), model.TypeIncome)

	if err := incomeQuery.Scan(&income).Error; err != nil {
		return 0, err
	}

	// Get total expense
	transactionQuery := db.conn.Table("transactions").
		Select("COALESCE(SUM(amount), 0) as total").
		Where("tg_id = ? AND date BETWEEN ? AND ? AND type = ?",
			tgID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), model.TypeExpense)

	if err := transactionQuery.Scan(&transaction).Error; err != nil {
		return 0, err
	}

	return income - transaction, nil
}

// GetMonthlyTotalsInYear gets monthly totals for a specific year
func (db *DB) GetMonthlyTotalsInYear(tgID int64, year int) (map[int]map[model.TransactionType]float64, error) {
	var results []struct {
		Month int
		Type  model.TransactionType
		Total float64
	}

	query := db.conn.Table("transactions").
		Select("EXTRACT(MONTH FROM date) as month, type, SUM(amount) as total").
		Where("tg_id = ? AND EXTRACT(YEAR FROM date) = ?", tgID, year).
		Group("month, type").
		Order("month")

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	// Convert to map of maps: month -> type -> amount
	monthlyTotals := make(map[int]map[model.TransactionType]float64)

	for _, result := range results {
		if _, exists := monthlyTotals[result.Month]; !exists {
			monthlyTotals[result.Month] = make(map[model.TransactionType]float64)
		}
		monthlyTotals[result.Month][result.Type] = result.Total
	}

	return monthlyTotals, nil
}

// GetUserTransactionsByMonthPaginated retrieves paginated transactions for a user for a specific year and month
func (db *DB) GetUserTransactionsByMonthPaginated(tgID int64, year int, month int, offset, limit int) ([]model.Transaction, int64, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1) // Last day of the month

	var transactions []model.Transaction
	var total int64

	// Get total count
	err := db.conn.Model(&model.Transaction{}).
		Where("tg_id = ? AND date BETWEEN ? AND ?",
			tgID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	result := db.conn.Where("tg_id = ? AND date BETWEEN ? AND ?",
		tgID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Order("date DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return transactions, total, nil
}

// GetUserTransactionsPaginated retrieves all transactions for a user with pagination
func (db *DB) GetUserTransactionsPaginated(tgID int64, offset, limit int) ([]model.Transaction, int64, error) {
	var transactions []model.Transaction
	var total int64

	// Get total count
	err := db.conn.Model(&model.Transaction{}).
		Where("tg_id = ?", tgID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	result := db.conn.Where("tg_id = ?", tgID).
		Order("date DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return transactions, total, nil
}

// SearchUserTransactions searches transactions by description with optional category filter
func (db *DB) SearchUserTransactions(tgID int64, searchQuery string, category string, offset, limit int) ([]model.Transaction, int64, error) {
	var transactions []model.Transaction
	var total int64

	// Build base query
	baseQuery := db.conn.Model(&model.Transaction{}).Where("tg_id = ?", tgID)

	// Add search condition (case-insensitive)
	baseQuery = baseQuery.Where("LOWER(description) LIKE LOWER(?)", "%"+searchQuery+"%")

	// Add category filter if not "all"
	if category != "" && category != "all" {
		baseQuery = baseQuery.Where("category = ?", category)
	}

	// Get total count
	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	result := baseQuery.
		Order("date DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return transactions, total, nil
}
