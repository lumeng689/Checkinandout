package tests

// func initTestFamilyCC(s controllers.CCServer) {
// 	// Create Institution for Tag-CC Test
// 	instToCreate := svc.InstitutionForm{
// 		Type:          string(svc.InstTypeSchool),
// 		MemberType:    string(svc.MemberTypeGuardian),
// 		WorkflowType:  string(svc.WorkflowTypeCC),
// 		Name:          "FAMILY_CC_TEST",
// 		Address:       "001 Test Drive",
// 		State:         "AZ",
// 		ZipCode:       "09999",
// 		RequireSurvey: false,
// 	}
// 	res, err := svc.CreateInst(instToCreate)
// 	if err != nil {
// 		panic(err)
// 	}
// 	// Create Family for Family-CC Test
// 	instID := res.InsertedID.(primitive.ObjectID).Hex()

// 	memberForm := svc.MemberInFamilyRegForm{
// 		PhoneNum:  "654-321-0987",
// 		Email:     "example1@123.com",
// 		FirstName: "John",
// 		LastName:  "Doe",
// 		Relation:  "Brother",
// 	}
// 	wardForm1 := svc.WardForm{
// 		FirstName: "Ward 1",
// 		LastName:  "Brown",
// 		Group:     "Level 1",
// 	}
// 	wardForm2 := svc.WardForm{
// 		FirstName: "Ward 2",
// 		LastName:  "Brown",
// 		Group:     "Level 2",
// 	}
// 	familyToCreate := svc.FamilyRegForm{
// 		InstID:  instID,
// 		Members: []svc.MemberInFamilyRegForm{memberForm},
// 		Wards:   []svc.WardForm{wardForm1, wardForm2},
// 	}

// _, err = svc.CreateFamily(familyToCreate)
// if err != nil {
// 	panic(err)
// }
// }
