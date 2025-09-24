package store

import (
	"database/sql"
	"go-learning/learn-gohttp/version4/internal/model"
)

// ProductStore 包含与商品相关的数据库操作
type ProductStore struct {
	db *sql.DB
}

// dbGetAllProducts 从数据库获取所有商品
func (s *ProductStore) dbGetAllProducts() ([]model.Product, error) {
	rows, err := s.db.Query("SELECT id, name, price, created_at FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
