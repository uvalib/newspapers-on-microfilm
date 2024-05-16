// include this on a lambda build only
//go:build lambda

package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func firstEntry(s []string) string {
	if len(s) == 0 {
		return ""
	}

	return s[0]
}

func proxyResponseGeneric(code int, contentType string, body string) events.APIGatewayProxyResponse {
	headers := make(map[string]string)

	if contentType != "" {
		headers["Content-Type"] = contentType
	}

	return events.APIGatewayProxyResponse{
		Headers:    headers,
		Body:       body,
		StatusCode: code,
	}
}

func proxyResponseError(code int, body string) events.APIGatewayProxyResponse {
	return proxyResponseGeneric(code, "text/plain", body)
}

func proxyResponseSuccess(body string) events.APIGatewayProxyResponse {
	return proxyResponseGeneric(http.StatusOK, "text/html", body)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var choice string
	var state string
	var year string
	var begin string
	var end string

	// collect form data
	switch request.HTTPMethod {

	case "GET":

		choice = request.QueryStringParameters["choice"]
		state = request.QueryStringParameters["file"]
		year = request.QueryStringParameters["year"]
		begin = request.QueryStringParameters["begin"]
		end = request.QueryStringParameters["end"]

	case "POST":

		post, err := url.ParseQuery(request.Body)

		if err != nil {
			return proxyResponseError(http.StatusInternalServerError, err.Error()), nil
		}

		choice = firstEntry(post["choice"])
		state = firstEntry(post["state"])
		year = firstEntry(post["year"])
		begin = firstEntry(post["begin"])
		end = firstEntry(post["end"])

	default:
		// invalid method
		return proxyResponseError(http.StatusMethodNotAllowed, fmt.Sprintf("unsupported method: [%s]", request.HTTPMethod)), nil
	}

	// perform lookup
	req := lookupRequest{choice: choice, state: state, year: year, begin: begin, end: end}
	res := lookup(req)

	if res.err != nil {
		return proxyResponseError(http.StatusInternalServerError, res.err.Error()), nil
	}

	// convert response to html
	page, err := res.toHTML()
	if err != nil {
		return proxyResponseError(http.StatusInternalServerError, err.Error()), nil
	}

	// success
	return proxyResponseSuccess(page), nil
}

func main() {
	lambda.Start(handler)
}
