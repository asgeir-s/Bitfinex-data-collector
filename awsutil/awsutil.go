package awsutil

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/cluda/btcdata/trade"
)

func SnsPublish(svc *sns.SNS, tick trade.Tick, topicArn string) (*sns.PublishOutput, error) {
	tickJSON, err := json.Marshal(tick)
	if err != nil {
		return nil, err
	}
	//last tick to SNS
	params := &sns.PublishInput{
		Message:          aws.String(string(tickJSON)), // Required
		MessageStructure: aws.String("messageStructure"),
		Subject:          aws.String("subject"),
		TopicArn:         aws.String(topicArn),
	}
	resp, err := svc.Publish(params)

	if err != nil {
		return nil, err
	}
	return resp, nil
}
