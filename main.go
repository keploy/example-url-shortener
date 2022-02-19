package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/keploy/go-sdk/integrations/kgin/v1"
	"github.com/keploy/go-sdk/integrations/kmongo"
	"github.com/keploy/go-sdk/keploy"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type url struct {
	ID      string    `json:"id" bson:"_id"`
	Created time.Time `json:"created" bson:"created"`
	Updated time.Time `json:"updated" bson:"updated"`
	URL     string    `json:"URL" bson:"url"`
}

var col *kmongo.Collection
var logger *zap.Logger

func Get(ctx context.Context, id string) (*url, error) {
	// too repetitive
	// TODO write a generic FindOne for all get calls
	filter := bson.M{"_id": id}
	var u url
	err := col.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func Upsert(ctx context.Context, u url) error {
	upsert := true
	opt := &options.UpdateOptions{
		Upsert: &upsert,
	}
	filter := bson.M{"_id": u.ID}
	update := bson.D{{"$set", u}}

	_, err := col.UpdateOne(ctx, filter, update, opt)
	if err != nil {
		return err
	}
	return nil
}

func getURL(c *gin.Context) {
	hash := c.Param("param")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please append url hash"})
		return
	}

	u, err := Get(c.Request.Context(), hash)
	if err != nil {
		logger.Error("failed to find url in the database", zap.Error(err), zap.String("hash", hash))
		c.JSON(http.StatusNotFound, gin.H{"error": "url not found"})
		return
	}
	c.Redirect(http.StatusSeeOther, u.URL)
	return
}

func putURL(c *gin.Context) {
	var m map[string]string

	err := c.ShouldBindJSON(&m)
	if err != nil {
		logger.Error("failed to decode req", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to decode req"})
		return
	}
	u := m["url"]

	if u == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing url param"})
		return
	}

	t := time.Now()
	id := GenerateShortLink(u)
	err = Upsert(c.Request.Context(), url{
		ID:      id,
		Created: t,
		Updated: t,
		URL:     u,
	})
	if err != nil {
		logger.Error("failed to save url to db", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": "http://localhost:8080/" + id})
}

func New(host, db string) (*mongo.Client, error) {
	clientOptions := options.Client()

	clientOptions.ApplyURI("mongodb://" + host + "/" + db + "?retryWrites=true&w=majority")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return mongo.Connect(ctx, clientOptions)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	logger, _ = zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	ddName, collection := "keploy", "url-shortener"
	client, err := New("localhost:27017", ddName)
	if err != nil {
		logger.Fatal("failed to create mgo db client", zap.Error(err))
	}
	db := client.Database(ddName)

	// integrate keploy with mongo
	col = kmongo.NewCollection(db.Collection(collection))

	port := "8080"
	// initialize keploy
	k := keploy.New(keploy.Config{
		App: keploy.AppConfig{
			Name: "sample-url-shortner",
			Port: "8080",
		},
		Server: keploy.ServerConfig{
			URL: "http://localhost:8081/api",
		},
	})

	r := gin.Default()

	// integrate keploy with gin router
	kgin.GinV1(k, r)

	r.GET("/:param", getURL)
	r.POST("/url", putURL)

	r.Run(":" + port)
}
