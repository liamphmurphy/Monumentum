package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type UserInfo struct {
	ShowName  string
	ShowDate  string
	ShowType  string
	UserEmail string
}

// Because the map from DB requires more than one value per key, this function returns multiple values per key
func MakeUserInfo(name string, showtype string, date string, email string) *UserInfo {
	return &UserInfo{
		ShowName:  name,
		ShowDate:  date,
		ShowType:  showtype,
		UserEmail: email,
	}
}

// This function only opens the db file using the sqlite3 driver, doesn't do anything else.
func InitializeDB() *sql.DB {
	db, err := sql.Open("sqlite3", "search.db")
	if err != nil {
		panic(err.Error())
	}
	return db
}

// When user hits submit, their values are added into the sqlite3 db.
func AddToDatabase(form url.Values) {
	db := InitializeDB()
	// Becase r.Form values is an array of strings, convert it to normal string for database purposes.
	showName := strings.Join(form["sname"], " ")
	showDate := strings.Join(form["sdate"], " ")
	userEmail := strings.Join(form["uemail"], " ")
	showType := strings.Join(form["showtype"], " ")

	// Useful information to print in console for server.
	fmt.Println("SHOW NAME: " + showName)
	fmt.Println("SHOW DATE: " + showDate)
	fmt.Println("USER EMAIL: " + userEmail)

	// Setup SQL query and then execute using user input values.
	insert, err := db.Prepare("INSERT INTO Reminders (ShowName, ShowType, ReleaseDate, Email) VALUES (?,?,?,?)")
	if err != nil {
		panic(err.Error())
	}
	insert.Exec(showName, showType, showDate, userEmail)
}

// Gets all current records in database and puts it into map.
func QueryDB() map[int]*UserInfo {
	db := InitializeDB()

	query, err := db.Query("SELECT ShowID, ShowName, ShowType, ReleaseDate, Email FROM Reminders")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	userInfo := make(map[int]*UserInfo)
	for query.Next() {
		var ShowName, ShowType, ReleaseDate, Email string
		var ShowID int
		query.Scan(&ShowID, &ShowName, &ShowType, &ReleaseDate, &Email)
		userInfo[ShowID] = MakeUserInfo(ShowName, ShowType, ReleaseDate, Email)

	}
	return userInfo

}
