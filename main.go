package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/cazanelena/book-app/register"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
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
	Isbn   string
}

var tpl *template.Template
var store = sessions.NewCookieStore([]byte("super-secret"))

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
	tpl = template.Must(template.ParseGlob("templates/*.html"))

	// Define routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/search", searchHandler(db))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/loginauth", func(w http.ResponseWriter, r *http.Request) {
		loginAuthHandler(w, r, db)
	})

	// Register the /registerauth route with a custom handler that has the database connection.
	http.HandleFunc("/registerauth", func(w http.ResponseWriter, r *http.Request) {
		register.RegisterAuthHandler(w, r, db)
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		register.RegisterHandler(w, r)
	})

	// Serve static files (CSS, JS, etc.)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the server
	fmt.Println("Server listening on port 8080")
	// http.Handle("/", r)
	// log.Fatal(http.ListenAndServe(":8080", nil))
	// // if you are not using gorilla/mux, you need to wrap your handler with context.ClearHandler
	http.ListenAndServe("localhost:8080", context.ClearHandler(http.DefaultServeMux))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprint(w, "Welcome to the Book Search!\nUse /search to find a book.")
	tpl.ExecuteTemplate(w, "search.html", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "login.html", nil)
}

func loginAuthHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("********loginAuthHandler running*******")
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Println("username:", username, "password:", password)

	// Check if the username (email) exists in the database
	var userID, hash string
	stmt := "SELECT hash FROM user_auth WHERE email = $1"
	row := db.QueryRow(stmt, username)
	err := row.Scan(&hash)
	if err == sql.ErrNoRows {
		// If the username does not exist, show a message asking the user to register first.
		fmt.Println("Username does not exist. Please register first.")
		tpl.ExecuteTemplate(w, "login.html", "Username does not exist in our database. Please register first.")
		return
	} else if err != nil {
		// Handle other errors that might occur during the database query.
		fmt.Println("Error selecting Hash in the db by Email/Username:", err)
		tpl.ExecuteTemplate(w, "login.html", "Something went wrong. Please try again later.")
		return
	}

	// Compare the hash with the password
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		// If the password is incorrect, show a message asking the user to check their login credentials.
		fmt.Println("Incorrect password. Please check your login credentials.")
		tpl.ExecuteTemplate(w, "login.html", "Incorrect password. Please check your login credentials.")
		return
	} else if err != nil {
		// Handle other errors that might occur during password comparison.
		fmt.Println("Error comparing hash with password:", err)
		tpl.ExecuteTemplate(w, "login.html", "Something went wrong. Please try again later.")
		return
	}
	// Get always returns a session, even if empty
	// Returns error if exists and coulld not be decoded
	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Session struct field Values map[interface{}]interface{}
	session.Values["userID"] = userID
	
	// Save before writing to response/return from handler
	session.Save(r, w)
	// If the password is correct, the user has successfully logged in.
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func logoutHandler (w http.ResponseWriter, r *http.Request) {
	fmt.Println("*********logoutHandler running*****")
	session, _ := store.Get(r, "session")
	// The delete built-in function delets the element with the specified key (m[key]) from the map
	// if m is nillor there is no such element, elete is a no-op
	delete(session.Values, "userId")
	session.Save(r, w)
	tpl.ExecuteTemplate(w, "login.html", "Logged Out")
}

func searchHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get session 
		session, _ := store.Get(r, "session")
		_, ok := session.Values["userID"]
		fmt.Println("ok", ok)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// Parse the query parameter from the URL
		query := r.URL.Query().Get("title")

		// Perform the database query
		rows, err := db.Query("SELECT title, authors, num_pages, average_rating, isbn FROM books WHERE title ILIKE '%' || $1 || '%'", query)
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
			err := rows.Scan(&bookData.Title, &bookData.Author, &bookData.Pages, &bookData.Rating, &bookData.Isbn)
			if err != nil {
				log.Printf("Error scanning query results: %v", err)
				http.Error(w, "Error processing query results", http.StatusInternalServerError)
				return
			}

			// Set the Found flag to true since a book is found
			bookData.Found = true

			books = append(books, bookData)
		}

		// Create a map to pass data to the template
		data := map[string]interface{}{
			"Books": books,
			"Query": query, // Pass the query string to the template
		}

		tpl.ExecuteTemplate(w, "search.html", data)
	}
}
