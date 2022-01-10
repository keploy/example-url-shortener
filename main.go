package main

import (
	"context"
	"encoding/json"
	"github.com/keploy/go-sdk/keploy"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"

	"github.com/bnkamalesh/webgo/v4"
	"github.com/bnkamalesh/webgo/v4/middleware/accesslog"
	"github.com/bnkamalesh/webgo/v4/middleware/cors"
	"github.com/keploy/go-sdk/integrations"
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

func getURL(w http.ResponseWriter, r *http.Request) {
	// WebGo context
	wctx := webgo.Context(r)
	// URI parameters, map[string]string
	hash := wctx.Params()["param"]
	if hash == "" {
		webgo.R404(
			w,
			map[string]string{
				"msg": "url not found",
			})
		return
	}

	u, err := Get(r.Context(), hash)
	if err != nil {
		logger.Error("failed to find url in the database", zap.Error(err), zap.String("hash", hash))
		webgo.R404(
			w,
			map[string]string{
				"msg": "url not found",
			})
		return
	}
	http.Redirect(w, r, u.URL, http.StatusSeeOther)
	return
}

func putURL(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var m map[string]string

	err := decoder.Decode(&m)
	if err != nil {
		logger.Error("failed to decode req", zap.Error(err))
		webgo.R500(
			w,
			map[string]string{
				"msg": "internal error",
			})
		return
	}
	u := m["url"]

	if u == "" {
		webgo.R404(
			w,
			map[string]string{
				"msg": "url not found",
			})
		return
	}

	t := time.Now()
	id := GenerateShortLink(u)
	err = Upsert(r.Context(), url{
		ID:      id,
		Created: t,
		Updated: t,
		URL:     u,
	})
	if err != nil {
		logger.Error("failed to save url to db", zap.Error(err))
		webgo.R500(
			w,
			map[string]string{
				"msg": "internal error",
			})
		return
	}

	webgo.R200(
		w,
		map[string]string{
			"url": "http://localhost:8080/" + id,
		})
}

func getRoutes() []*webgo.Route {
	return []*webgo.Route{
		{
			Name:          "post-url",
			Method:        http.MethodPost,
			Pattern:       "/url",
			Handlers:      []http.HandlerFunc{putURL},
			TrailingSlash: true,
		},
		{
			Name:                    "get-url",
			Method:                  http.MethodGet,
			Pattern:                 "/:param",
			Handlers:                []http.HandlerFunc{getURL},
			TrailingSlash:           true,
			FallThroughPostResponse: true,
		},
	}
}

var col *integrations.MongoDB
var logger *zap.Logger

func New(host, db string) (*mongo.Client, error) {
	clientOptions := options.Client()

	clientOptions.ApplyURI("mongodb://" + host + "/" + db + "?retryWrites=true&w=majority")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return mongo.Connect(ctx, clientOptions)
}

func main() {
	logger, _ = zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	ddName, collection := "keploy", "url-shortener"
	client, err := New("localhost:27017", ddName)
	if err != nil {
		logger.Fatal("failed to create mgo db client", zap.Error(err))
	}
	db := client.Database(ddName)
	col = integrations.NewMongoDB(db.Collection(collection))

	cfg := &webgo.Config{
		Host:         "",
		Port:         "8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	router := webgo.NewRouter(cfg, getRoutes())

	router.UseOnSpecialHandlers(accesslog.AccessLog)
	router.Use(accesslog.AccessLog)
	router.Use(cors.CORS(nil))
	webgo.GlobalLoggerConfig(
		nil, nil,
		webgo.LogCfgDisableDebug,
	)
	kply := keploy.NewApp("url-shortener", "<API_KEY>", "https://api.keploy.io", "0.0.0.0", "8080")
	integrations.WebGoV4(kply, router)
	router.Start()
}
