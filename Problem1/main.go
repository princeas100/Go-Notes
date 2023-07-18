package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type Note struct {
	ID   uint32 `json:"id"`
	Note string `json:"note"`
}

type Details struct {
	Name     string
	Email    string
	Password string
}

var Detail = []Details{
	{"prince", "prince@gmail.com", "as@123"},
}

var UserNotes = make(map[string][]Note)

var Sessions = make(map[string]string)

func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser Details
	err := json.NewDecoder(r.Body).Decode(&newUser)

	if err != nil {
		http.Error(w, "Invalid request data.", http.StatusBadRequest)
		return
	}

	blankFields := []string{}

	if newUser.Email == "" {
		blankFields = append(blankFields, "email")
	} else if !isValidEmail(newUser.Email) {
		http.Error(w, "Invalid email format. Please provide a valid email address.", http.StatusBadRequest)
		return
	}

	if newUser.Name == "" {
		blankFields = append(blankFields, "name")
	}

	if newUser.Password == "" {
		blankFields = append(blankFields, "password")
	}

	if len(blankFields) > 0 {
		errorMessage := fmt.Sprintf("Missing required fields: %s.", strings.Join(blankFields, ", "))
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	Detail = append(Detail, newUser)
	w.WriteHeader(http.StatusOK)
}
func loginUser(w http.ResponseWriter, r *http.Request) {
	var loginUser Details
	err := json.NewDecoder(r.Body).Decode(&loginUser)

	if err != nil {
		http.Error(w, "Invalid request data.", http.StatusBadRequest)
		return
	}

	blankFields := []string{}

	if loginUser.Email == "" {
		blankFields = append(blankFields, "email")
	} else if !isValidEmail(loginUser.Email) {
		http.Error(w, "Invalid email format. Please provide a valid email address.", http.StatusBadRequest)
		return
	}

	if loginUser.Password == "" {
		blankFields = append(blankFields, "password")
	}

	if len(blankFields) > 0 {
		errorMessage := fmt.Sprintf("Missing required fields: %s.", strings.Join(blankFields, ", "))
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	for _, user := range Detail {
		if user.Email == loginUser.Email && user.Password == loginUser.Password {
			sessionID := generateSessionID()

			Sessions[loginUser.Email] = sessionID

			response := map[string]string{"message": "Login successful", "sid": sessionID}
			responseJSON, err := json.Marshal(response)
			if err != nil {
				http.Error(w, "Failed to respond.", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseJSON)

			return
		}
	}

	http.Error(w, "Username and/or password do not match.", http.StatusUnauthorized)

	w.WriteHeader(http.StatusOK)

}

func listUserNotes(w http.ResponseWriter, r *http.Request) {

	var requestBody struct {
		SID string `json:"sid"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request data.", http.StatusBadRequest)
		return
	}


	userEmail, validSession := validateSession(requestBody.SID)
	if !validSession {
		http.Error(w, "Invalid session ID.", http.StatusUnauthorized)
		return
	}


	notes, found := UserNotes[userEmail]
	if !found {
		notes = generateDummyNotes()
	}


	response := map[string][]Note{"notes": notes}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to respond.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func createNote(w http.ResponseWriter, r *http.Request) {

	var requestBody struct {
		SID  string `json:"sid"`
		Note string `json:"note"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request data.", http.StatusBadRequest)
		return
	}

	userEmail, validSession := validateSession(requestBody.SID)
	if !validSession {
		http.Error(w, "Invalid session ID.", http.StatusUnauthorized)
		return
	}


	newNoteID := generateUniqueNoteID()
	if newNoteID == 0 {
		http.Error(w, "Failed to generate a unique note ID.", http.StatusInternalServerError)
		return
	}


	newNote := Note{ID: newNoteID, Note: requestBody.Note}


	UserNotes[userEmail] = append(UserNotes[userEmail], newNote)

	response := map[string]uint32{"id": newNoteID}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal response.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		SID string `json:"sid"`
		ID  uint32 `json:"id"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request data.", http.StatusBadRequest)
		return
	}

	userEmail, validSession := validateSession(requestBody.SID)
	if !validSession {
		http.Error(w, "Invalid session ID.", http.StatusUnauthorized)
		return
	}

	notes, found := UserNotes[userEmail]
	if !found {
		http.Error(w, "User has no notes.", http.StatusBadRequest)
		return
	}

	index := -1
	for i, note := range notes {
		if note.ID == requestBody.ID {
			index = i
			break
		}
	}

	if index == -1 {
		http.Error(w, "Note not found.", http.StatusBadRequest)
		return
	}

	UserNotes[userEmail] = append(notes[:index], notes[index+1:]...)

	w.WriteHeader(http.StatusOK)
}


func generateUniqueNoteID() uint32 {

	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return 0
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func validateSession(sessionID string) (string, bool) {
	for email, sid := range Sessions {
		if sid == sessionID {
			return email, true
		}
	}
	return "", false
}

func generateDummyNotes() []Note {
	dummyNotes := []Note{
		{ID: 1, Note: "This is the first dummy note."},
		{ID: 2, Note: "This is the second dummy note."},
		{ID: 3, Note: "This is the third dummy note."},
	}

	return dummyNotes
}

func generateSessionID() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(emailRegex).MatchString(email)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/signup", createUser).Methods("POST")
	router.HandleFunc("/login", loginUser).Methods("POST")
	router.HandleFunc("/notes", listUserNotes).Methods("GET")
	router.HandleFunc("/notes", createNote).Methods("POST")
	router.HandleFunc("/notes", deleteNote).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}
