/*
 * Copyright 2024 Kapeta Inc.
 * SPDX-License-Identifier: MIT
 */
package pubsub

import (
	"encoding/json"
	"errors"
	"github.com/kapetacom/schemas/packages/go/model"
	"github.com/kapetacom/sdk-go-config/providers"
)

func getTopic(config providers.ConfigProvider, resourceName string) (string, error) {
	consumer, err := getConsumer(config, resourceName)
	if err != nil {
		return "", nil
	}

	if consumer == nil {
		return "", errors.New("No consumer found for " + resourceName)
	}

	return consumer.Spec.Topic, nil
}

func getConsumer(config providers.ConfigProvider, resourceName string) (*PubSubProviderConsumer, error) {
	provider, err := config.GetInstancesForProvider(resourceName)
	if err != nil {
		return nil, err
	}

	if len(provider) == 0 || len(provider[0].Connections) == 0 {
		return nil, errors.New("No connection found for " + resourceName)
	}

	consumerResourceName := provider[0].Connections[0].Consumer.ResourceName

	blockDefinition, err := toPubSubBlockDefinition(provider[0].Block)
	if err != nil {
		return nil, err
	}

	if blockDefinition.Spec.Consumers == nil {
		return nil, errors.New("No consumer found for " + resourceName)
	}

	for _, consumer := range blockDefinition.Spec.Consumers {
		if consumer.Metadata.Name == consumerResourceName {
			return &consumer, nil
		}
	}

	return nil, nil
}

func toPubSubBlockDefinition(block *model.Kind) (*PubSubBlockDefinition, error) {
	j, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}
	v := &PubSubBlockDefinition{}
	err = json.Unmarshal(j, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}
