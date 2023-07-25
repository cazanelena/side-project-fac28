package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
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

var templates = "./templates/"

func main() {
	// Connecting to the database
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer db.Close()

	// Initialize the route
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/search", searchHandler(db))

	// Serve static files (CSS, JS, etc.)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the server
	fmt.Println("Server listening on port 8080")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprint(w, "Welcome to the Book Search!\nUse /search to find a book.")
	tpl := template.Must(template.ParseFiles(templates + "search.html"))
	tpl.Execute(w, nil)
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


        tpl := template.Must(template.ParseFiles(templates + "search.html"))

        // Create a map to pass data to the template
        data := map[string]interface{}{
            "Books": books,
            "Query": query, // Pass the query string to the template
        }
		
        tpl.Execute(w, data)
    }
}
