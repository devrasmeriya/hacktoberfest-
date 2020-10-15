package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Meeting struct {
	ID                string          `json:"id,omitempty" validate:"required" bson:"id,omitempty"`
	Title             string          `json:"title,omitempty" validate:"required" bson:"title,omitempty"`
	Participants      []*Participants `json:"participants,omitempty" validate:"required" bson:"participants,omitempty"`
	StartTime         string          `json:"starttime,omitempty" validate:"required" bson:"starttime,omitempty"`
	EndTime           string          `json:"endtime,omitempty" validate:"required" bson:"endtime,omitempty"`
	CreationTimestamp string          `json:"timestamp,omitempty" bson:"Timestamp,omitempty"`
}

type Participants struct {
	Name  string `json:"name,omitempty" validate:"required" bson:"name,omitempty"`
	Email string `json:"email,omitempty" validate:"required,email" bson:"email,omitempty"`
	RSVP  string `json:"rsvp,omitempty" validate:"required" bson:"rsvp,omitempty"`
}

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/meetings", CreateMeetingEndpoint).Methods("POST")
	router.HandleFunc("/meetings/{id}", GetMeetingEndpoint).Methods("GET")
	router.HandleFunc("/meetings?start={starttime}&end={endtime}", GetTimeEndpoint).Methods("GET")
	router.HandleFunc("/articles?participant={email}", GetEmailEndpoint).Methods("GET")
	http.ListenAndServe(":1800", router)
}

func CreateMeetingEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var meetings Meeting
	_ = json.NewDecoder(request.Body).Decode(&meetings)
	meetings.ID = strconv.Itoa(rand.Intn(10000))

	collection := client.Database("MeetingAPI").Collection("MeetingDetails")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	collection.InsertOne(ctx, meetings)
	json.NewEncoder(response).Encode(meetings)
}

func GetMeetingEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := (params["id"])
	var meetings Meeting
	collection := client.Database("MeetingAPI").Collection("MeetingDetails")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, Meeting{ID: id}).Decode(&meetings)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(meetings)
}
func GetEmailEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var participant []Meeting
	params := mux.Vars(request)
	email, _ := (params["email"])
	collection := client.Database("thepolyglotdeveloper").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{"email": email})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Meeting
		cursor.Decode(&person)
		participant = append(participant, person)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(participant)
}

func GetTimeEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var meetings []Meeting
	params := mux.Vars(request)
	starttime, _ := (params["starttime"])
	endtime, _ := (params["endtime"])
	collection := client.Database("thepolyglotdeveloper").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{"starttime": starttime, "endtime": endtime})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var meeting Meeting
		cursor.Decode(&meeting)
		meetings = append(meetings, meeting)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(meetings)
}
