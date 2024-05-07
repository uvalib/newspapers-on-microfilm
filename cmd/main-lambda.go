// include this on a lambda build only
//go:build lambda

package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// collect form data
	choice := request.QueryStringParameters["choice"]
	state := request.QueryStringParameters["file"]
	year := request.QueryStringParameters["year"]
	begin := request.QueryStringParameters["begin"]
	end := request.QueryStringParameters["end"]

	// perform lookup
	req := lookupRequest{choice: choice, state: state, year: year, begin: begin, end: end}
	res := lookup(req)

	// convert response to html
	buf, err := res.toHTML()

	// if error, return 500 with error text as body
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       res.err.Error(),
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
