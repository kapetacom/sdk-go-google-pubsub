/*
 * Copyright 2024 Kapeta Inc.
 * SPDX-License-Identifier: MIT
 */
package pubsub

import "github.com/kapetacom/schemas/packages/go/model"

type PubSubBlockDefinition struct {
	Kind     string          `json:"kind"`
	Metadata model.Metadata  `json:"metadata"`
	Spec     PubSubBlockSpec `json:"spec"`
}

type PubSubBlockSpec struct {
	Providers []PubSubProviderConsumer `json:"providers"`
	Consumers []PubSubProviderConsumer `json:"consumers"`
}

type PubSubProviderConsumer struct {
	Metadata model.ResourceMetadata      `json:"metadata"`
	Spec     PubSubTopicSubscriptionSpec `json:"spec"`
}

type PubSubTopicSubscriptionSpec struct {
	Topic        string `json:"topic"`
	Subscription string `json:"subscription"`
}
