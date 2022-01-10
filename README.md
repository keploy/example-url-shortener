# test-app-url-shortner
A sample url shortener app to test Keploy integration capabilities

The app does 2 things: 
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
To add the keploy the SDK we need to wrap the dynamodb client and webgo router. Instructions are here - https://github.com/keploy/go-sdk/blob/main/README.md

You can also checkout to the keploy branch which has the integrations already done and enter your API key in the keploy method
```go
	kply := keploy.NewApp("url-shortener", "<API_KEY>", "https://api.keploy.io", host, port)
```

## Capture mode
To capture testcases, set the `KEPLOY_SDK_MODE` to "capture" and start the app
```bash
export KEPLOY_SDK_MODE="capture" && go run generator.go main.go
```

Now perform any of the above requests, and they will be captured as a testcase and would be visible in the web UI
![testcases](testcases.png?raw=true "Web console testcases")

## Test mode
Now that we have our testcase captured, we can run them. We need to change the `KEPLOY_SDK_MODE` to 'test`
```bash
export KEPLOY_SDK_MODE="test" && go run generator.go main.go
```
**Now you can also stop mongo. It'll be mocked by the SDK!**

In about 5 secs delay, the tests would start running. The logs will help us understand the status. We can get deeper insight through the test runs tab in the web console. 

![testruns](testrun1.png?raw=true "Recent testruns")
![testruns](testrun2.png?raw=true "Summary")
![testruns](testrun3.png?raw=true "Detail")

