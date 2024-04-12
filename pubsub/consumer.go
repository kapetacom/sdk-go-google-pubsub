/*
 * Copyright 2024 Kapeta Inc.
 * SPDX-License-Identifier: MIT
 */
package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kapetacom/sdk-go-config/providers"
)

type MessageHandler[T any, Attributes map[string]string] func(message T, Attributes map[string]string) error

type Consumer[T any] struct {
	handler      MessageHandler[T, map[string]string]
	subscription *pubsub.Subscription
}

func CreateConsumer[T any, Attributes map[string]string](config providers.ConfigProvider, resourceName string, handler MessageHandler[T, map[string]string]) (*Consumer[T], error) {
	instance, err := config.GetInstanceForConsumer(resourceName)
	if err != nil {
		return nil, err
	}

	blockSpec, err := toPubSubBlockDefinition(instance.Block)
	if err != nil {
		return nil, fmt.Errorf("error decoding block spec: %v", err)
	}

	topicName, subscriptionName, err := getTopicSubscription(blockSpec, resourceName)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, pubsub.DetectProjectID)
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

	subscription := client.Subscription(subscriptionName)

	exists, err = subscription.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if !exists {
		subscription, err = client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic: topic,
		})
	}

	return &Consumer[T]{
		subscription: subscription,
		handler:      handler,
	}, nil
}

func getTopicSubscription(spec *PubSubBlockDefinition, resourceName string) (string, string, error) {
	var provider PubSubTopicSubscriptionSpec

	for _, b := range spec.Spec.Providers {
		if b.Metadata.Name == resourceName {
			provider = b.Spec
			break
		}
	}

	if provider == (PubSubTopicSubscriptionSpec{}) {
		return "", "", errors.New("topic subscription not found for resource " + resourceName)
	}

	return provider.Topic, provider.Subscription, nil
}

func (c *Consumer[T]) ReceiveMessages(ctx context.Context) {
	cancelContext, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-cancelContext.Done():
			return
		default:
			err := c.subscription.Receive(cancelContext, func(ctx context.Context, msg *pubsub.Message) {
				value, err := c.unmarshalValue(msg.Data)
				if err != nil {
					msg.Nack()
					return
				}

				err = c.handler(value, msg.Attributes)
				if err != nil {
					msg.Nack()
					return
				}

				msg.Ack()
			})

			if err != context.Canceled {
				// todo: error handling
			}
		}
	}
}

func (c *Consumer[T]) unmarshalValue(data []byte) (T, error) {
	var value T
	err := json.Unmarshal(data, &value)
	if err != nil {
		return value, err
	}
	return value, nil
}
