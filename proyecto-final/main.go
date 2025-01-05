
package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Category string  `json:"category"`
}

var templates = template.Must(template.ParseGlob("templates/*.html"))
var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./inventory.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable()
	populateTable()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/add", addProductHandler)
	http.HandleFunc("/delete", deleteProductHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Web client running on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		price REAL,
		quantity INTEGER,
		category TEXT
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func populateTable() {
	query := `
	INSERT INTO products (name, price, quantity, category) VALUES
	('Tomato Seeds', 10.50, 100, 'Seeds'),
	('Fertilizer', 15.00, 50, 'Soil'),
	('Greenhouse Frame', 500.00, 5, 'Equipment')
	ON CONFLICT DO NOTHING;
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM products")
	if err != nil {
		http.Error(w, "Error fetching products", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity, &p.Category); err != nil {
			http.Error(w, "Error scanning products", http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	if err := templates.ExecuteTemplate(w, "index.html", products); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func addProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	product := Product{
		Name:     r.FormValue("name"),
		Price:    parseFloat(r.FormValue("price")),
		Quantity: parseInt(r.FormValue("quantity")),
		Category: r.FormValue("category"),
	}

	_, err := db.Exec("INSERT INTO products (name, price, quantity, category) VALUES (?, ?, ?, ?)",
		product.Name, product.Price, product.Quantity, product.Category)
	if err != nil {
		http.Error(w, "Error adding product", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	productID := parseInt(r.FormValue("id"))

	_, err := db.Exec("DELETE FROM products WHERE id = ?", productID)
	if err != nil {
		http.Error(w, "Error deleting product", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func parseFloat(value string) float64 {
	v, _ := strconv.ParseFloat(value, 64)
	return v
}

func parseInt(value string) int {
	v, _ := strconv.Atoi(value)
	return v
}
