# test-app-url-shortner
A sample url shortener app to test Keploy integration capabilities

## Installation
```bash
git clone https://github.com/keploy/test-app-url-shortner
cd test-app-url-shortner
go mod tidy
```
The App also requires mongo
```bash
docker container run -it  -p 27017:27017 mongo
```

## Add Keploy SDK
To add the keploy the SDK we need to wrap the dependencies of the url-shortner app, here, `dynamodb client` and `webgo router`.

See complete instructions to integrate Keploy Go SDK at [keploy/go-sdk](https://github.com/keploy/go-sdk/blob/main/README.md)

For demo purpose, checkout to the [keploy branch](https://github.com/keploy/test-app-url-shortner/tree/keploy) which has the integrations done already.

```bash
git checkout keploy 
```

Now enter your API key in the keploy initialization method in `main.go`.

```go
kply := keploy.NewApp("url-shortener", "<API_KEY>", "https://api.keploy.io", host, port)
```

## Capture mode
To capture testcases, set the `KEPLOY_SDK_MODE` environment variable to `capture` and run the app

```bash
export KEPLOY_SDK_MODE="capture" && go run generator.go main.go
```

Let's **capture some traffic** by making some API calls. You can use [Postman](https://www.postman.com/), [Hoppscotch](https://hoppscotch.io/), or simply `curl`

1. Generate shortned url

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

2. Redirect to original url from shortened url
```bash
curl --request GET \
  --url http://localhost:8080/Lhr4BWAi
```

or by querying through the browser `http://localhost:8080/Lhr4BWAi`


Now both these API calls were captured as a testcase and should be visible on the **Keploy console**.
If you're using Keploy cloud, open [Console](https://app.keploy.io/testlist).

You should be seeing an app named url-shortener with the test cases we just captured.

![testcases](testcases.png?raw=true "Web console testcases")


Now, let's see the magic! 🪄💫


## Test mode

Now that we have our testcase captured. **Shut down your mongo docker container. 😐**

Change the `KEPLOY_SDK_MODE` to `test`
```bash
export KEPLOY_SDK_MODE="test" && go run generator.go main.go
```

**Run the application again!**

All the test-cases will be downloaded locally and run with the app.


Guess what!?


**MongoDB calls are mocked while running these test cases!**

So no need to setup dependencies like
mongoDB, web-go locally during testing.

**The application thinks it's talking to
mongoDB 😄**

Go to the Keploy Console/testruns to get deeper insights on what testcases ran, what failed.

![testruns](testrun1.png?raw=true "Recent testruns")
![testruns](testrun2.png?raw=true "Summary")
![testruns](testrun3.png?raw=true "Detail")

If you are using other dependencies, feel free to create an issue, we'll add the support asap! 