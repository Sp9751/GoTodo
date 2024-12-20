package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Completed bool `json:completed` 
	Body string `json:body`
}

var collection *mongo.Collection

func main (){
	app:= fiber.New()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error loading .env:", err)
	}
	PORT := os.Getenv("PORT")
	DB_URI := os.Getenv("DB_URI")

	clientOptions := options.Client().ApplyURI(DB_URI)

	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("connected to mongodb")

	collection = client.Database("goTodo").Collection("todos")

	app.Get("/api/todo", getTodos)
	app.Post("/api/todo", createTodo)
	app.Patch("/api/todo/:id", updateTodo)
	app.Delete("/api/todo/:id", deleteTodo)

	app.Listen(":"+PORT)
}

func getTodos (c *fiber.Ctx)error {
	var todos []Todo

	cursor , err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()){
		var todo Todo

		if err := cursor.Decode(&todo); err != nil {
			return err
		}

		todos = append(todos, todo)
	}

	return c.Status(200).JSON(todos)
}
func createTodo (c *fiber.Ctx)error {
	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Body == ""{
		return c.Status(400).JSON(fiber.Map{"error": "Todo body can't be empty"})
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)

	if err != nil {
		return err
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(200).JSON(todo)
}

func updateTodo (c *fiber.Ctx)error {
	id := c.Params("id")
	ObjectID ,err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"Error": "Invalid todo ID"})
	}

	filter := bson.M{"_id": ObjectID}
	update := bson.M{"$set":bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"Success": true})

}

func deleteTodo (c *fiber.Ctx)error {
	id:= c.Params("id")

	ObjectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"Error": "Invalid todo ID"})
	}

	filter := bson.M{"_id": ObjectID}

	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"Success": true})
}