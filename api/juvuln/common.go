// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package juvuln

type SubscriberScheme string
type NewSubscriberScheme SubscriberScheme

// Defines values for SubscriberScheme.
const (
	SubscriberSchemeHttp SubscriberScheme = "http"

	SubscriberSchemeHttps SubscriberScheme = "https"
)
