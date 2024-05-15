// include this on a lambda build only
//go:build lambda

package main

import (
	"fmt"
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
			return events.APIGatewayProxyResponse{
				Body:       err.Error(),
				StatusCode: 500,
			}, nil
		}

		choice = firstEntry(post["choice"])
		state = firstEntry(post["state"])
		year = firstEntry(post["year"])
		begin = firstEntry(post["begin"])
		end = firstEntry(post["end"])

	default:
		// invalid method
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("unsupported method: [%s]", request.HTTPMethod),
			StatusCode: 405,
		}, nil
	}

	// perform lookup
	req := lookupRequest{choice: choice, state: state, year: year, begin: begin, end: end}
	res := lookup(req)

	// convert response to html
	buf, err := res.toHTML()

	// if error, return 500 with error text as body
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, nil
	}

	// success
	return events.APIGatewayProxyResponse{
		Body:       buf,
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
