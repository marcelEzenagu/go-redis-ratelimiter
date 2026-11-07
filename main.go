package main

import (
	"context"
	"net/http"
	"os"
	limiter "rate-limiter/middleware"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func main() {
	// http-server
	router := gin.Default()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// redis
	rDB := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx := context.Background()
	_, err := rDB.Ping(ctx).Result()
	if err != nil {
		panic("Redis not connected")
	}

	limitP := limiter.NewLimiter(rDB)
	// route
	router.GET("/api", func(c *gin.Context) {

		key := c.ClientIP()
		limit := 4

		// middleware
		allowed, remains, err := limitP.Allow(c, key, limit, time.Minute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("X-Rate-Limit-Remaining", strconv.FormatInt(remains, 10))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"message": "too many request"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "request allowed"})

	})

	router.GET("/metrics", func(c *gin.Context) {
		reqCtx := c.Request.Context()
		ip := c.Query("ip")
		if ip == "" {
			ip = c.ClientIP()
		}

		count, err := limitP.GetCount(reqCtx, ip)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ip": ip, "count": count})

	})

	router.Run(":8083")
}
