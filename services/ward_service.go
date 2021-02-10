package services

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetNewWard - as is
func GetNewWard(w WardForm) Ward {
	return Ward{
		ID:        primitive.NewObjectID(),
		FirstName: w.FirstName,
		LastName:  w.LastName,
		Group:     w.Group,
	}
}
