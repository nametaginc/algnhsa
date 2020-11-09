package algnhsa

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

var (
	errALBUnexpectedRequest         = errors.New("expected ALBTargetGroupRequest")
	errALBExpectedMultiValueHeaders = errors.New("expected multi value headers; enable Multi value headers in target group settings")
)

func getALBSourceIP(event events.ALBTargetGroupRequest) string {
	if xff, ok := event.MultiValueHeaders["x-forwarded-for"]; ok && len(xff) > 0 {
		ips := strings.SplitN(xff[0], ",", 2)
		if len(ips) > 0 {
			return ips[0]
		}
	}
	return ""
}

func newALBRequest(ctx context.Context, payload []byte, opts *Options) (lambdaRequest, error) {
	var event events.ALBTargetGroupRequest
	if err := json.Unmarshal(payload, &event); err != nil {
		return lambdaRequest{}, err
	}
	if event.RequestContext.ELB.TargetGroupArn == "" {
		return lambdaRequest{}, errALBUnexpectedRequest
	}
	if len(event.MultiValueHeaders) == 0 {
		return lambdaRequest{}, errALBExpectedMultiValueHeaders
	}



	req := lambdaRequest{
		HTTPMethod:                      event.HTTPMethod,
		Path:                            event.Path,
		QueryStringParameters:           map[string]string{},
		MultiValueQueryStringParameters: map[string][]string{},
		Headers:                         event.Headers,
		MultiValueHeaders:               event.MultiValueHeaders,
		Body:                            event.Body,
		IsBase64Encoded:                 event.IsBase64Encoded,
		SourceIP:                        getALBSourceIP(event),
		Context:                         newTargetGroupRequestContext(ctx, event),
	}

	for k, v:= range event.QueryStringParameters {
		k, err := url.QueryUnescape(k)
		if err != nil {
			return lambdaRequest{}, err
		}
		v, err := url.QueryUnescape(v)
		if err != nil {
			return lambdaRequest{}, err
		}
		req.QueryStringParameters[k] = v
	}
	for k, vv:= range event.MultiValueQueryStringParameters {
		k, err := url.QueryUnescape(k)
		if err != nil {
			return lambdaRequest{}, err
		}
		for i, v := range vv {
			vv[i], err = url.QueryUnescape(v)
			if err != nil {
				return lambdaRequest{}, err
			}
		}
		req.MultiValueQueryStringParameters[k] = vv
	}



	return req, nil
}
