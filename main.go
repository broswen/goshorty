package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/ksuid"
)

type ShortRequest struct {
	Url  string `json:"url" binding:"required"`
	Slug string `json:"slug"`
}

var ctx = context.Background()
var rdb *redis.Client

const (
	timeoutSeconds = 10
)

func pingHandler(c *gin.Context) {
	c.String(200, "Pong")
}

func shortHandler(c *gin.Context) {
	// bind request into
	request := ShortRequest{}
	if err := c.ShouldBind(&request); err != nil {
		log.Println(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("%+v", request)
	slug, err := getSlug(request.Slug)

	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	err = rdb.Set(ctx, slug, request.Url, time.Second*timeoutSeconds).Err()
	if err != nil {
		panic(err)
	}

	c.JSON(200, map[string]string{
		"slug": slug,
	})
}

func getSlug(slug string) (string, error) {

	if slug == "" {
		id, err := ksuid.NewRandom()
		if err != nil {
			return "", err
		}
		return id.String(), nil
	}

	val, err := rdb.Get(ctx, slug).Result()
	log.Println("existing slug", val)
	if err != nil && err != redis.Nil {
		return "", err
	}
	if err == redis.Nil {
		return slug, nil
	}
	return "", errors.New("slug already exists")
}

func slugHandler(c *gin.Context) {
	url, err := rdb.Get(ctx, c.Param("slug")).Result()
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Redirect(http.StatusMovedPermanently, url)
}

func init() {
	log.Println("connecting redis client")
	rdb = redis.NewClient(&redis.Options{
		Addr:     "redis-server:6379",
		Password: "",
		DB:       0,
	})
}

func main() {
	r := gin.Default()

	r.GET("/ping", pingHandler)
	r.POST("/short", shortHandler)
	r.GET("/:slug", slugHandler)

	r.Run()
}
