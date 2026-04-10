package config

import (
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	BooksCollection        = "books"
	DepartmentsCollection  = "departments"
	SpecialitiesCollection = "specialities"
	PatientsCollection     = "patients"
	ShiftsCollection       = "shifts"
	StaffCollection        = "staff"
)

type Collections struct {
	Books        *mongo.Collection
	Departments  *mongo.Collection
	Specialities *mongo.Collection
	Patients     *mongo.Collection
	Shifts       *mongo.Collection
	Staff        *mongo.Collection
}

func GetCollections(client *mongo.Client, dbName string) *Collections {
	mongoDB := client.Database(dbName)

	return &Collections{
		Books:        mongoDB.Collection(BooksCollection),
		Departments:  mongoDB.Collection(DepartmentsCollection),
		Specialities: mongoDB.Collection(SpecialitiesCollection),
		Patients:     mongoDB.Collection(PatientsCollection),
		Shifts:       mongoDB.Collection(ShiftsCollection),
		Staff:        mongoDB.Collection(StaffCollection),
	}
}
func GetCollection(client *mongo.Client, dbName string, collectionName string) *mongo.Collection {
	return client.Database(dbName).Collection(collectionName)
}
