package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type NotificationSubscription struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Topic string             `bson:"topic" json:"topic"`
	Email string             `bson:"email" json:"email"`
}
