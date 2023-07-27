package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	_ "github.com/lib/pq"
	_ "github.com/go-sql-driver/mysql"
	"github.com/cazanelena/book-app/register"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "secret"
	dbname   = "book_db"
)

type BookData struct {
	Found  bool
	Title  string
	Author string
	Pages  int
	Rating float64
}


var tpl *template.Template

func main() {
	// Connecting to the database
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer db.Close()

	// Check the database connection
	if err = db.Ping(); err != nil {
		log.Fatal("Error pinging the database: ", err)
	}

	// Initialize the route
	tpl = template.Must(template.ParseGlob("templates/*html"))

	// Define routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/search", searchHandler(db))
	http.HandleFunc("/login", loginHandler)

	// Register the /registerauth route with a custom handler that has the database connection.
	http.HandleFunc("/registerauth", func(w http.ResponseWriter, r *http.Request) {
		register.RegisterAuthHandler(w, r, db)
	})

	// Serve static files (CSS, JS, etc.)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the server
	fmt.Println("Server listening on port 8080")
	// http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprint(w, "Welcome to the Book Search!\nUse /search to find a book.")
	tpl.ExecuteTemplate(w, "search.html", nil)
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "login.html", nil)
}

func searchHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query parameter from the URL
		query := r.URL.Query().Get("title")

		// Perform the database query
		rows, err := db.Query("SELECT title, authors, num_pages, average_rating FROM books WHERE title ILIKE '%' || $1 || '%'", query)
		if err != nil {
			log.Printf("Database query error: %v", err)
			http.Error(w, "Database query error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Process the query results
		var books []BookData

		for rows.Next() {
			var bookData BookData
			err := rows.Scan(&bookData.Title, &bookData.Author, &bookData.Pages, &bookData.Rating)
			if err != nil {
				log.Printf("Error scanning query results: %v", err)
				http.Error(w, "Error processing query results", http.StatusInternalServerError)
				return
			}

			// Set the Found flag to true since a book is found
			bookData.Found = true

			books = append(books, bookData)
		}

		tpl := template.Must(template.ParseFiles("./templates/search.html"))

		// Create a map to pass data to the template
		data := map[string]interface{}{
			"Books": books,
			"Query": query, // Pass the query string to the template
		}

		tpl.Execute(w, data)
	}
}

