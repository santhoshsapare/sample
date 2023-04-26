package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"google.golang.org/api/option"
	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

type Employee struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

func main() {
	// Set up credentials
	ctx := context.Background()
	credsFile := "/home/saba/gcp-key/credentials.json"
	opt := option.WithCredentialsFile(credsFile)


	// Initialize Firestore client
	ctx = context.Background()
	client, err := firestore.NewClient(ctx, "develop-375210", opt)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	// Initialize router
	router := mux.NewRouter()

	// Get all employees
	router.HandleFunc("/employees", func(w http.ResponseWriter, r *http.Request) {
		iter := client.Collection("employees").Documents(ctx)
		var employees []Employee
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatalf("Failed to iterate over employees: %v", err)
			}
			var employee Employee
			if err := doc.DataTo(&employee); err != nil {
				log.Fatalf("Failed to decode employee data: %v", err)
			}
			employees = append(employees, employee)
		}
		json.NewEncoder(w).Encode(employees)
	}).Methods("GET")

	// Get an employee by ID
	router.HandleFunc("/employees/{id}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		doc, err := client.Collection("employees").Doc(id).Get(ctx)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Employee not found")
			return
		}
		var employee Employee
		if err := doc.DataTo(&employee); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to decode employee data")
			return
		}
		json.NewEncoder(w).Encode(employee)
	}).Methods("GET")

	// Create a new employee
	router.HandleFunc("/employees", func(w http.ResponseWriter, r *http.Request) {
		var employee Employee
		json.NewDecoder(r.Body).Decode(&employee)
		docRef, _, err := client.Collection("employees").Add(ctx, employee)
		if err != nil {
			log.Fatalf("Failed to add employee to Firestore: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Employee created with ID: %s", docRef.ID)
	}).Methods("POST")

	// Update an employee
	router.HandleFunc("/employees/{id}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		var employee Employee
		json.NewDecoder(r.Body).Decode(&employee)
		_, err := client.Collection("employees").Doc(id).Set(ctx, employee)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to update employee")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Employee updated successfully")
	}).Methods("PUT")

	// Delete an employee
	router.HandleFunc("/employees/{id}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		_, err := client.Collection("employees").Doc(id).Delete(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to delete employee")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Employee deleted successfully")
	}).Methods("DELETE")

	// Start server
	log.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
