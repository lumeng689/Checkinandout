package services

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SurveyRegForm struct {
	InstID   string           `bson:"institution_id" json:"institution_id"`
	MemberID string           `bson:"member_id" json:"member_id"`
	QAList   []QuestionAnswer `json:"qa_list"`
}

type QuestionAnswer struct {
	QuestionIndex   string  `bson:"question_index" json:"question_index"`
	Question        string  `bson:"question" json:"question"`
	AnswerBool      bool    `bson:"answer_bool" json:"answer_bool"`
	AnswerNumerical float64 `bson:"answer_numerical" json:"answer_numerical"`
	AnswerText      string  `bson:"answer_text" json:"answer_text"`
}

type Survey struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	InstID    string             `bson:"institution_id" json:"institution_id"`
	MemberID  string             `bson:"member_id" json:"member_id"`
	QAList    []QuestionAnswer   `json:"qa_list"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type GetSurveyParams struct {
	InstID string `json:"inst_id"`
}

var surveyCollection *mongo.Collection

func SurveyCollection(c *mongo.Database) {
	surveyCollection = c.Collection("surveys")
}

func GetManySurveys(params *GetSurveyParams) (*mongo.Cursor, error) {
	var filters bson.D
	filters = append(filters, primitive.E{Key: "institution_id", Value: params.InstID})
	return surveyCollection.Find(context.TODO(), filters)
}

func CreateSurvey(s SurveyRegForm, qas []QuestionAnswer) (*mongo.InsertOneResult, error) {

	newSurvey := Survey{
		ID:        primitive.NewObjectID(),
		InstID:    s.InstID,
		MemberID:  s.MemberID,
		QAList:    qas,
		CreatedAt: time.Now(),
	}
	return surveyCollection.InsertOne(context.TODO(), newSurvey)
}
