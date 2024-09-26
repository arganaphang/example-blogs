package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlogSQL struct {
	ID        uint      `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

const (
	TABLE_NAME             = "blogs"
	MATERIALIZE_TABLE_NAME = "random_blogs"
)

type BlogMongo struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Title     string             `json:"title" bson:"title"`
	Content   string             `json:"content" bson:"content"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type Application struct {
	sqlDB   *sqlx.DB
	mongoDB *mongo.Database
	// mongoDB
}

func main() {
	app := fiber.New()
	app.Use(logger.New())

	// Initiate Database
	// SQL
	postgresURL, ok := os.LookupEnv("POSTGRES_URL")
	if !ok {
		log.Fatalln("POSTGRES_URL is missing")
	}
	sqlDB, err := sqlx.Open("postgres", postgresURL)
	if err != nil {
		log.Fatalln("failed to connect to postgres", err.Error())
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalln("failed to ping postgres", err.Error())
	}
	// Mongo
	mongoURL, ok := os.LookupEnv("MONGO_URL")
	if !ok {
		log.Fatalln("MONGO_URL is missing")
	}
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalln("failed to connect to mongo", err.Error())
	}
	if err := mongoClient.Ping(context.TODO(), nil); err != nil {
		log.Fatalln("failed to ping mongo", err.Error())
	}

	application := Application{
		sqlDB:   sqlDB,
		mongoDB: mongoClient.Database("test"),
	}

	// Health
	app.Get("/healthz", application.healthz)

	// SQL
	sql := app.Group("/sql")
	sql.Get("/blog", application.sqlBlog)
	sql.Get("/random", application.sqlRandom)

	mongo := app.Group("mongo")
	mongo.Get("/blog", application.mongoBlog)
	mongo.Get("/random", application.mongoRandom)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-c
		fmt.Println("Gracefully shutting down...")
		_ = app.Shutdown()
		_ = sqlDB.Close()
		_ = mongoClient.Disconnect(context.TODO())
	}()

	// Scheduler
	// jakarta, _ := time.LoadLocation("Asia/Jakarta")
	// scheduler := cron.New(cron.WithLocation(jakarta))
	// defer scheduler.Stop()
	// scheduler.AddFunc("@daily", func() { // */5 * * * * <- every 5 minutes crontab.guru
	// 	_, err := application.sqlDB.Exec(fmt.Sprintf("REFRESH MATERIALIZED VIEW %s", MATERIALIZE_TABLE_NAME))
	// 	if err != nil {
	// 		logrus.Info("failed to randomize blogs")
	// 		logrus.Info(err.Error())
	// 		return
	// 	}
	// })
	// go scheduler.Start()

	app.Listen("0.0.0.0:8000")
}

func (a Application) healthz(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(map[string]string{
		"message": "OK",
	})
}

func (a Application) sqlBlog(c *fiber.Ctx) error {
	limitQuery := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": "failed to parse limit query",
		})
	}
	offsetQuery := c.Query("offset", "0")
	offset, err := strconv.Atoi(offsetQuery)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": "failed to parse offset query",
		})
	}
	blogs := []BlogSQL{}
	if err := a.sqlDB.Select(&blogs, fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", MATERIALIZE_TABLE_NAME), limit, offset); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": "failed to get blogs",
		})
	}
	return c.Status(http.StatusOK).JSON(map[string]any{
		"message": "get sql blog posts",
		"data":    blogs,
	})
}

func (a Application) sqlRandom(c *fiber.Ctx) error {
	if _, err := a.sqlDB.Exec(fmt.Sprintf("REFRESH MATERIALIZED VIEW %s", MATERIALIZE_TABLE_NAME)); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": "failed to randomize blogs",
			"error":   err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(map[string]string{
		"message": "seed successfully",
	})
}

func (a Application) mongoBlog(c *fiber.Ctx) error {
	limitQuery := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": "failed to parse limit query",
		})
	}
	offsetQuery := c.Query("offset", "0")
	offset, err := strconv.Atoi(offsetQuery)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{
			"message": "failed to parse offset query",
		})
	}
	csr, err := a.mongoDB.Collection("blogs").Find(
		c.Context(),
		bson.M{},
		options.
			Find().
			SetLimit(int64(limit)).
			SetSkip(int64(offset)),
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer csr.Close(c.Context())

	blogs := make([]BlogMongo, 0)
	for csr.Next(c.Context()) {
		var row BlogMongo
		err := csr.Decode(&row)
		if err != nil {
			log.Fatal(err.Error())
		}
		blogs = append(blogs, row)
	}

	return c.Status(http.StatusOK).JSON(map[string]any{
		"message": "get mongo blog posts",
		"data":    blogs,
	})
}

func (a Application) mongoRandom(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(map[string]string{
		"message": "seed successfully",
	})
}
