package register
import (
	"fmt"
	"unicode"
	"net/http"
	"golang.org/x/crypto/bcrypt"
	"time"
	"html/template"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"

)

var tpl *template.Template

func RegisterAuthHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Println("*********registerAuthHandler running*****")
	r.ParseForm()
	username := r.FormValue("username")

	// check username has only alphaNumeric characters
	var nameAlphaNumeric = true

	tpl = template.Must(template.ParseGlob("templates/*html"))

	for _, char := range username {
		if unicode.IsLetter(char) == false && unicode.IsNumber(char) == false {
			nameAlphaNumeric = false
		}
	}

	// check username length
	var nameLength bool
	if 5 <= len(username) && len(username) <= 50 {
		nameLength = true
	}

	// check password criteria
	password := r.FormValue("password")
	fmt.Println("passord:", password, "\npswdLength:", len(password))

	var pswdLowercase, pswdUppercase, pswdNumber, pswdSpecial, pswdLength, pswdNoSpace bool
	pswdNoSpace = true

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			pswdLowercase = true
		case unicode.IsUpper(char):
			pswdUppercase = true
		case unicode.IsNumber(char):
			pswdNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			pswdSpecial = true
		case unicode.IsSpace(int32(char)):
			pswdNoSpace = false
		}
	}
	if 11 < len(password) && len(password) < 50 {
		pswdLength = true
	}
	fmt.Println("pswdLowercase:", pswdLowercase, "\npswdUppercase:", pswdUppercase, "\npswdNumber:", pswdNumber, "\npswdSpecial:", pswdSpecial, "\npswdLength:", pswdLength, "\npswdNoSpaces:", pswdNoSpace, "\nnameAlphaNumeric:", nameAlphaNumeric, "\nnameLength:", nameLength)
	if !pswdLowercase || !pswdUppercase || !pswdNumber || !pswdSpecial || !pswdLength || !pswdNoSpace || !nameAlphaNumeric || !nameLength {
		tpl.ExecuteTemplate(w, "register.html", "please check username and password criteria")
		return
	}

	// check if username already exists for availability
	stmt := "SELECT id FROM user_auth WHERE email = $1"
	row := db.QueryRow(stmt, username)
	var uID string
	err := row.Scan(&uID)
	if err != sql.ErrNoRows {
		fmt.Println("username already exists, err:", err)
		tpl.ExecuteTemplate(w, "register.html", "username already taken")
		return
	}

	// create hash from password
	var hash []byte
	// func GenerateFromPassword(password []byte, cost int) ([]byte, error)
	hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("bcrypt err:", err)
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	fmt.Println("hash:", hash)
	fmt.Println("string(hash):", string(hash))
	// func (db *DB) Prepare(query string) (*Stmt, error)
	var insertStmt *sql.Stmt
	insertStmt, err = db.Prepare("INSERT INTO user_auth (email, hash, created_on) VALUES ($1, $2, $3);")
	if err != nil {
		fmt.Println("error preparing statement:", err)
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	defer insertStmt.Close()
	createdOn := time.Now()
	createdOnStr := createdOn.Format("2006-01-02 15:04:05")

	println(createdOnStr)
	var result sql.Result

	result, err = insertStmt.Exec(username, string(hash), createdOnStr)
	if err != nil {
		fmt.Println("error preparing statement:", err)
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	rowsAff, _ := result.RowsAffected()
	lastIns, _ := result.LastInsertId()
	fmt.Println("rowsAff:", rowsAff)
	fmt.Println("lastIns:", lastIns)
	fmt.Println("err:", err)

	if err != nil {
		fmt.Println("error inserting new user")
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	fmt.Fprint(w, "congrats, your account has been successfully created")
}
