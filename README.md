# Example URL Shortener
A sample url shortener app to test Keploy integration capabilities

## Installation
```bash
git clone https://github.com/keploy/example-url-shortener
cd example-url-shortener
go mod tidy
```
The App also requires mongo
```bash
docker container run -it  -p 27017:27017 mongo
```

## Add Keploy SDK
To add the keploy the SDK we need to wrap the dependencies of the url-shortner app, here, `dynamodb client` and `webgo router`. See complete instructions to integrate Keploy Go SDK at [keploy/go-sdk](https://github.com/keploy/go-sdk/blob/main/README.md)

If you'd like to check out the application without the keploy sdk you can switch to [this branch](https://github.com/keploy/example-url-shortener/tree/without-keploy).

### Initialize Keploy
```go
k := keploy.New(keploy.Config{
    App: keploy.AppConfig{
        Name: "sample-url-shortner",
        Port: "8080",
    },
    Server: keploy.ServerConfig{
        URL: "http://localhost:8081/api",
    },
})
```
### Integrate router
In this example we are using the gin router. To integrate with the gin router
```go
kgin.GinV1(k, r)
```

### Integrate Database
Likewise in this example we are using mongodb. To integrate with mongodb
```go
col = kmongo.NewCollection(db.Collection(collection))
```

And thats it!🔥

## Generate testcases
### Run the application
```shell
go run generator.go main.go
```

To genereate testcases we just need to make some API calls. You can use [Postman](https://www.postman.com/), [Hoppscotch](https://hoppscotch.io/), or simply `curl`

###1. Generate shortned url

```bash
curl --request POST \
  --url http://localhost:8080/url \
  --header 'content-type: application/json' \
  --data '{
  "url": "https://google.com"
}'
```
this will return the shortned url
```
{
  "data": {
    "url": "http://localhost:8080/Lhr4BWAi"
  },
  "status": 200
}
```

###2. Redirect to original url from shortened url
```bash
curl --request GET \
  --url http://localhost:8080/Lhr4BWAi
```

or by querying through the browser `http://localhost:8080/Lhr4BWAi`


Now both these API calls were captured as a testcase and should be visible on the **Keploy console**.
If you're using Keploy cloud, open [Console](https://app.keploy.io/testlist).

You should be seeing an app named `sample-url-shortner` with the test cases we just captured.

![testcases](testcases.png?raw=true "Web console testcases")


Now, let's see the magic! 🪄💫


## Test mode

Now that we have our testcase captured. **Shut down your mongo docker container. 😐**

### Method 1

Change the `KEPLOY_SDK_MODE` to `test` and **Run the application again!**
```bash
export KEPLOY_SDK_MODE="test" && go run generator.go main.go
```

### Method 2
You can use go-test to instrument tests and also calculate code coverage. 
```go
// main_test.go

package main

import (
	"github.com/keploy/go-sdk/keploy"
	"testing"
)

func TestKeploy(t *testing.T) {
	keploy.SetTestMode()
	go main()
	keploy.AssertTests(t)
}
```
then run the test file
```shell
 go test -coverpkg=./... -covermode=atomic  ./...
```
output should look like
```shell
ok      test-app-url-shortner   6.265s  coverage: 77.1% of statements in ./...
```

All the test-cases will be downloaded locally and run with the app, and we got 77% coverage without writing any piece of code. 


And Guess what!?


**MongoDB calls are mocked while running these test cases!**

So no need to setup dependencies like mongoDB, web-go locally or write mocks for your testing.

**The application thinks it's talking to
mongoDB 😄**

Go to the Keploy Console/testruns to get deeper insights on what testcases ran, what failed.

![testruns](testrun1.png?raw=true "Recent testruns")
![testruns](testrun2.png?raw=true "Summary")
![testruns](testrun3.png?raw=true "Detail")

If you are using other dependencies, feel free to create an issue, we'll add the support asap! 
