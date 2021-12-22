package app

import (
	"database/sql"
	"log"
)

type Product struct {
	Id    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func (p *Product) getProduct(db *sql.DB) error {
	return db.QueryRow("SELECT name, price FROM product WHERE id=$1",
		p.Id).Scan(&p.Name, &p.Price)
}
func (p *Product) updateProduct(db *sql.DB) error {
	_, err := db.Exec("UPDATE PRODUCT WITH(ROWLOCK) SET name=$1, price=$2 WHERE id=$3", p.Name, p.Price, p.Id)
	return err
}

func (p *Product) deleteProduct(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM PRODUCT WHERE id=$1", p.Id)
	return err
}
func (p *Product) createProduct(db *sql.DB) error {
	err := db.QueryRow("INSERT INTO PRODUCT(name,price) OUTPUT Inserted.id values($1,$2) ", p.Name, p.Price).Scan(&p.Id)
	if err != nil {
		return err
	}
	return nil
}

func getProducts(db *sql.DB, count int) ([]Product, error) {
	rows, err := db.Query("select  id, name,price from product with(nolock)")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.Id, &p.Name, &p.Price); err != nil {
			log.Fatal("error while iterating over rows", err)
		}
		products = append(products, p)
	}
	if rows.Err() != nil {
		log.Fatal("error while iterating over rows", rows.Err())
	}
	rows.Close()
	return products, nil
}
