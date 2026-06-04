/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// UnavailableOfferings stores offerings that recently failed to launch because of host-capacity
// exhaustion. Offerings present in the cache are treated as unavailable (Available=false) when
// instance types are listed, so the scheduler routes around them (including spot->on-demand and
// cross-NodePool fallback) until the entry's TTL expires.
type UnavailableOfferings struct {
	cache *cache.Cache
	ttl   time.Duration
}

// NewUnavailableOfferings creates the cache with the given TTL for how long an offering observed to
// be out of host capacity stays unavailable before Karpenter retries it. A non-positive ttl falls
// back to UnavailableOfferingsTTL.
func NewUnavailableOfferings(ttl time.Duration) *UnavailableOfferings {
	if ttl <= 0 {
		ttl = UnavailableOfferingsTTL
	}
	return &UnavailableOfferings{
		cache: cache.New(ttl, UnavailableOfferingsCleanupInterval),
		ttl:   ttl,
	}
}

// MarkUnavailable records the given offering as unavailable for the configured TTL. Calling it
// again for an already-cached offering refreshes the TTL.
func (u *UnavailableOfferings) MarkUnavailable(ctx context.Context, shape, zone, capacityType string) {
	log.FromContext(ctx).WithValues(
		"shape", shape,
		"zone", zone,
		"capacity-type", capacityType,
		"ttl", u.ttl,
	).V(1).Info("marking offering as unavailable")
	u.cache.SetDefault(u.key(shape, zone, capacityType), struct{}{})
}

// IsUnavailable returns true if the offering is currently cached as unavailable.
func (u *UnavailableOfferings) IsUnavailable(shape, zone, capacityType string) bool {
	_, found := u.cache.Get(u.key(shape, zone, capacityType))
	return found
}

func (u *UnavailableOfferings) Flush() {
	u.cache.Flush()
}

// key returns the cache key for an offering. Format: <capacityType>:<shape>:<zone>.
func (u *UnavailableOfferings) key(shape, zone, capacityType string) string {
	return fmt.Sprintf("%s:%s:%s", capacityType, shape, zone)
}
