package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/vynquoc/cs-flash-cards/internal/data"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	s3 struct {
		region     string
		bucketName string
	}
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
	s3     *s3.S3
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var cfg config
	flag.StringVar(&cfg.env, "env", "development", "Environment")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("CSFLASHCARDS_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-mx-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	cfg.s3.region = os.Getenv("AWS_REGION")
	cfg.s3.bucketName = os.Getenv("S3_BUCKET")
	port, err := strconv.Atoi(os.Getenv("PORT"))

	if err != nil {
		logger.Fatal(err)
	}
	cfg.port = port

	db, err := openDB(cfg)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Printf("database connection pool established")

	s3Session, err := createS3session(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		s3:     s3Session,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 & time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)

}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func createS3session(cfg config) (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.s3.region),
	})
	if err != nil {
		return nil, err
	}
	return s3.New(sess), nil
}
