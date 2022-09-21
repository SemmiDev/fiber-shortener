package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	db "github.com/SemmiDev/fiber-shortener/db/sqlc"
	"github.com/SemmiDev/fiber-shortener/util"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/teris-io/shortid"
)

type server struct {
	config     util.Config
	store      db.Store
	app        *fiber.App
	redisStore *redis.Client
	redisQueue *util.RedisQueue
	generator  *shortid.Shortid
}

type request struct {
	URL string `json:"url"`
}

func (s *server) shortenURLHandler(c *fiber.Ctx) error {
	var req request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cannot parse body")
	}

	host := string(c.Context().Host())

	// check if url already exist in redis
	shortURL, err := s.redisStore.Get(c.Context(), req.URL).Result()
	if err == nil {
		return c.JSON(fiber.Map{
			"short_url": host + "/" + shortURL,
		})
	}

	link, err := s.store.CreateLinkTx(c.Context(), db.CreateLinkParams{
		LongUrl:  req.URL,
		ShortUrl: s.generator.MustGenerate(),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "cannot create short url")
	}

	go s.redisQueue.Enqueue(
		util.KVRedisQueue{Key: link.LongUrl, Val: link.ShortUrl},
		util.KVRedisQueue{Key: link.ShortUrl, Val: link.LongUrl},
	)
	return c.JSON(fiber.Map{
		"short_url": host + "/" + shortURL,
	})
}

func (s *server) resolveURLHandler(c *fiber.Ctx) error {
	shortURL := c.Params("short_url")
	if shortURL == "" {
		return fiber.NewError(fiber.StatusBadRequest, "please provide short url")
	}

	longURL, err := s.redisStore.Get(c.Context(), shortURL).Result()
	if err == nil {
		return c.Redirect(longURL)
	}

	link, err := s.store.GetLinkByShortURL(c.Context(), shortURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return fiber.NewError(fiber.StatusNotFound, "cannot find short url")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "cannot get short url")
	}

	go s.redisQueue.Enqueue(
		util.KVRedisQueue{Key: link.LongUrl, Val: link.ShortUrl},
		util.KVRedisQueue{Key: link.ShortUrl, Val: link.LongUrl},
	)

	return c.Redirect(link.LongUrl)
}

func main() {
	// load config from app.env file
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// connect to database
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// ping to db
	if err := conn.Ping(); err != nil {
		log.Fatal("cannot ping db:", err)
	}

	//set connection pool
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// run migration
	migration, err := migrate.New(config.MigrationURL, config.DBSource)
	if err != nil {
		log.Fatal("cannot create new migrate instance:", err)
	}
	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to run migrate up:", err)
	}

	// create new store
	store := db.NewStore(conn)

	// create redis client and connect to redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "secret",
		DB:       0,
	})

	err = redisClient.Ping(context.Background()).Err()
	if err != nil {
		log.Fatal("cannot connect to redis:", err)
	}

	redisQueue := util.NewRedisQueue(redisClient)

	// create generator instance
	generator, err := shortid.New(31, shortid.DefaultABC, 2342)
	if err != nil {
		log.Fatal("cannot create new generator instance:", err)
	}

	// create new fiber instance
	app := fiber.New()

	// set middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// create new server
	s := &server{
		app:        app,
		config:     config,
		store:      store,
		redisStore: redisClient,
		redisQueue: redisQueue,
		generator:  generator,
	}

	s.app.Static("/", "./frontend/src")

	// register routes
	s.app.Post("/shorten", s.shortenURLHandler)
	s.app.Get("/:short_url", s.resolveURLHandler)

	// start server
	err = s.app.Listen(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start http server:", err)
	}
}
