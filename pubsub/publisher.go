/*
 * Copyright 2024 Kapeta Inc.
 * SPDX-License-Identifier: MIT
 */
package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"github.com/kapetacom/sdk-go-config/providers"
)

type Publisher[DataType any, Headers map[string]string] struct {
	topic *pubsub.Topic
	client *pubsub.Client
}

type PublisherPayload[DataType any, Headers map[string]string] struct {
	DataType DataType          `json:"data"`
	Headers  map[string]string `json:"headers"`
}

func CreatePublisher[DataType any, Headers map[string]string](config providers.ConfigProvider, resourceName string) (*Publisher[DataType, Headers], error) {

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, pubsub.DetectProjectID)
	if err != nil {
		return nil, err
	}

	topicName, err := getTopic(config, resourceName)
	if err != nil {
		return nil, err
	}

	topic := client.Topic(topicName)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if !exists {
		topic, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			return nil, err
		}
	}

	return &Publisher[DataType, Headers]{
		topic: topic,
		client: client,
	}, nil
}

func (p *Publisher[DataType, Headers]) Publish(payload PublisherPayload[DataType, Headers]) (string, error) {
	payloadJSON, err := json.Marshal(payload.DataType)
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	res := p.topic.Publish(ctx, &pubsub.Message{
		Attributes: payload.Headers,
		Data:       payloadJSON,
	})

	serverID, err := res.Get(context.Background())
	if err != nil {
		return "", err
	}

	return serverID, nil
}

func (p *Publisher[DataType, Headers]) Close() error {
	return p.client.Close()
}