package main

import (
	"html/template" // Manejo de plantillas HTML
	"log"           // Registro de mensajes y errores
	"net/http"      // Manejo de solicitudes HTTP
	"strconv"       // Conversión de cadenas a números

	"database/sql" // Manejo de bases de datos SQL

	_ "github.com/mattn/go-sqlite3" // Driver para SQLite
)

// Product define la estructura de un producto en la base de datos.
type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Category string  `json:"category"`
}

// PageData representa los datos necesarios para las vistas con paginación.
type PageData struct {
	Products    []Product
	CurrentPage int
	TotalPages  int
}

var templates *template.Template // Plantillas HTML para renderizar vistas
var db *sql.DB                   // Conexión a la base de datos SQLite

// Inicializa la base de datos
func main() {
	var err error
	db, err = sql.Open("sqlite3", "./inventory.db")
	if err != nil {
		log.Fatal(err) // Detiene la ejecución si hay un error
	}

	// Agregar funciones personalizadas a las plantillas
	templates = template.Must(template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b }, // Suma dos enteros
		"sub": func(a, b int) int { return a - b }, // Resta dos enteros
	}).ParseGlob("templates/*.html"))

	// Crea la tabla y popula con datos iniciales
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

// homeHandler gestiona la página principal con el listado de productos y paginación.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	const pageSize = 5 // Número de productos por página
	totalProducts := getTotalProducts()
	totalPages := (totalProducts + pageSize - 1) / pageSize

	if page > totalPages && totalPages > 0 {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	rows, err := db.Query("SELECT * FROM products LIMIT ? OFFSET ?", pageSize, (page-1)*pageSize)
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

	data := PageData{
		Products:    products,
		CurrentPage: page,
		TotalPages:  totalPages,
	}

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// getTotalProducts obtiene el número total de productos en la base de datos.
func getTotalProducts() int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		log.Printf("Error counting products: %v", err)
		return 0
	}
	return count
}

// addProductHandler maneja la adición de productos.
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

// deleteProductHandler maneja la eliminación de productos.
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

// parseFloat convierte una cadena a float64.
func parseFloat(value string) float64 {
	v, _ := strconv.ParseFloat(value, 64)
	return v
}

// parseInt convierte una cadena a int.
func parseInt(value string) int {
	v, _ := strconv.Atoi(value)
	return v
}
