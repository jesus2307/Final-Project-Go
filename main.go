package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

type Article struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Content string `json:"content"`
}

var db *sql.DB

func main() {
	// Initialize the database
	initDatabase()

	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/articles", handleArticles)
	http.HandleFunc("/articles/", handleArticleByID)

	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func initDatabase() {
	var err error
	db, err = sql.Open("sqlite", "articles.db")
	if err != nil {
		panic(err)
	}

	createTable := `CREATE TABLE IF NOT EXISTS articles (
		id INTEGER PRIMARY KEY,
		title TEXT,
		desc TEXT,
		content TEXT
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	html := "<!DOCTYPE html>" +
		"<html lang=\"en\">" +
		"<head>" +
		"<meta charset=\"UTF-8\">" +
		"<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">" +
		"<title>Articles Management</title>" +
		"<style>" +
		"body { font-family: Arial, sans-serif; margin: 20px; }" +
		"table { width: 100%; border-collapse: collapse; margin-top: 20px; }" +
		"th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }" +
		"th { background-color: #f4f4f4; }" +
		"form { margin-top: 20px; }" +
		"</style>" +
		"</head>" +
		"<body>" +
		"<h1>Articles Management</h1>" +
		"<div id=\"articles-container\">" +
		"<h2>All Articles</h2>" +
		"<table id=\"articles-table\">" +
		"<thead>" +
		"<tr>" +
		"<th>ID</th>" +
		"<th>Title</th>" +
		"<th>Description</th>" +
		"<th>Content</th>" +
		"<th>Actions</th>" +
		"</tr>" +
		"</thead>" +
		"<tbody></tbody>" +
		"</table>" +
		"</div>" +
		"<div id=\"add-article-form\">" +
		"<h2>Add New Article</h2>" +
		"<form id=\"article-form\">" +
		"<label for=\"id\">ID:</label><br>" +
		"<input type=\"text\" id=\"id\" name=\"id\" required><br><br>" +
		"<label for=\"title\">Title:</label><br>" +
		"<input type=\"text\" id=\"title\" name=\"title\" required><br><br>" +
		"<label for=\"desc\">Description:</label><br>" +
		"<input type=\"text\" id=\"desc\" name=\"desc\" required><br><br>" +
		"<label for=\"content\">Content:</label><br>" +
		"<textarea id=\"content\" name=\"content\" required></textarea><br><br>" +
		"<button type=\"submit\">Add Article</button>" +
		"</form>" +
		"</div>" +
		"<script>" +
		"const apiUrl = '/articles';" +
		"async function fetchArticles() {" +
		"const response = await fetch(apiUrl);" +
		"const articles = await response.json();" +
		"const tbody = document.querySelector('#articles-table tbody');" +
		"tbody.innerHTML = '';" +
		"articles.forEach(article => {" +
		"const row = document.createElement('tr');" +
		"row.innerHTML = '<td>' + article.id + '</td>' +" +
		"'<td>' + article.title + '</td>' +" +
		"'<td>' + article.desc + '</td>' +" +
		"'<td>' + article.content + '</td>' +" +
		"'<td><button onclick=\"deleteArticle(' + article.id + ')\">Delete</button></td>';" +
		"tbody.appendChild(row);" +
		"});" +
		"}" +
		"document.getElementById('article-form').addEventListener('submit', async (e) => {" +
		"e.preventDefault();" +
		"const id = document.getElementById('id').value;" +
		"const title = document.getElementById('title').value;" +
		"const desc = document.getElementById('desc').value;" +
		"const content = document.getElementById('content').value;" +
		"await fetch(apiUrl, {" +
		"method: 'POST'," +
		"headers: { 'Content-Type': 'application/json' }," +
		"body: JSON.stringify({ id: parseInt(id), title, desc, content })" +
		"});" +
		"fetchArticles();" +
		"e.target.reset();" +
		"});" +
		"async function deleteArticle(id) {" +
		"await fetch(apiUrl + '/' + id, { method: 'DELETE' });" +
		"fetchArticles();" +
		"}" +
		"fetchArticles();" +
		"</script>" +
		"</body>" +
		"</html>"

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleArticles(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rows, err := db.Query("SELECT id, title, desc, content FROM articles")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var articles []Article
		for rows.Next() {
			var article Article
			if err := rows.Scan(&article.ID, &article.Title, &article.Desc, &article.Content); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			articles = append(articles, article)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(articles)
		return
	}

	if r.Method == http.MethodPost {
		var newArticle Article
		if err := json.NewDecoder(r.Body).Decode(&newArticle); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("INSERT INTO articles (id, title, desc, content) VALUES (?, ?, ?, ?)", newArticle.ID, newArticle.Title, newArticle.Desc, newArticle.Content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleArticleByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/articles/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodDelete {
		_, err := db.Exec("DELETE FROM articles WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
