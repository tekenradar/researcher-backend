package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type ParticipantContact struct {
	ID              primitive.ObjectID        `bson:"_id,omitempty" json:"id,omitempty"`
	AddedAt         int64                     `bson:"addedAt" json:"addedAt"`
	SessionID       string                    `bson:"sessionID" json:"sessionID"`
	ParticipantID   string                    `bson:"participantID" json:"participantID"`
	KeepContactData bool                      `bson:"keepContactData" json:"keepContactData"`
	General         ContactDetailsGeneralData `bson:"general" json:"general"`
	ContactData     ContactDetailsContactData `bson:"contactData" json:"contactData"`
	Notes           []ContactNote             `bson:"notes" json:"notes"`
}

type ContactDetailsGeneralData struct {
	Age          int    `bson:"age" json:"age"`
	Gender       string `bson:"string" json:"string"`
	OtherStudies bool   `bson:"otherStudies" json:"otherStudies"`
}

type ContactDetailsContactData struct {
	FirstName string   `bson:"firstName" json:"firstName"`
	LastName  string   `bson:"lastName" json:"lastName"`
	Birthday  int64    `bson:"birthday" json:"birthday"`
	Email     string   `bson:"email" json:"email"`
	Phone     string   `bson:"phone" json:"phone"`
	Gender    string   `bson:"gender" json:"gender"`
	GP        *GPInfos `bson:"gp" json:"gp"`
}

type GPInfos struct {
	Office  string  `bson:"office" json:"office"`
	Name    string  `bson:"name" json:"name"`
	Address Address `bson:"address" json:"address"`
	Phone   string  `bson:"phone" json:"phone"`
}

type Address struct {
	Street   string `bson:"street" json:"street"`
	Nr       string `bson:"nr" json:"nr"`
	Postcode string `bson:"postcode" json:"postcode"`
	City     string `bson:"city" json:"city"`
}

type ContactNote struct {
	ID      string `bson:"id" json:"id"`
	Time    int64  `bson:"time" json:"time"`
	Author  string `bson:"author" json:"author"`
	Content string `bson:"content" json:"content"`
}
