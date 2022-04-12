package middleware

import (
	"fmt"

	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//Log to file
func LoggerToFile() gin.HandlerFunc {
	src, err := os.OpenFile("wallet.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	// instantiation
	logger := logrus.New()

	//Set output
	logger.Out = src

	//Set log level
	logger.SetLevel(logrus.DebugLevel)

	//Format log
	logger.SetFormatter(&logrus.TextFormatter{})

	return func(c *gin.Context) {
		//Start time
		startTime := time.Now()

		//Process request
		c.Next()

		//End time
		endTime := time.Now()

		//Execution time
		latencyTime := endTime.Sub(startTime)

		//Request method
		reqMethod := c.Request.Method

		//Request routing
		reqUri := c.Request.RequestURI

		// status code
		statusCode := c.Writer.Status()

		// request IP
		clientIP := c.ClientIP()

		//Log format
		logger.Infof("| %3d | %13v | %s | %s | %s",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
		)
	}
}

//Log to mongodb
func LoggerToMongo() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

//Log to es
func LoggerToES() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

//Logging to MQ
func LoggerToMQ() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
