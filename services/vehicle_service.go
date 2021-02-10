package services

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetNewVehicle - as is
func GetNewVehicle(w VehicleForm) Vehicle {
	return Vehicle{
		ID:       primitive.NewObjectID(),
		Make:     w.Make,
		Model:    w.Model,
		Color:    w.Color,
		PlateNum: w.PlateNum,
	}
}
