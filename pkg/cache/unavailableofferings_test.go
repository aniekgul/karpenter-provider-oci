/*
** Karpenter Provider OCI
**
** Copyright (c) 2026 Oracle and/or its affiliates.
** Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/
 */

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnavailableOfferings_MarkAndIsUnavailable(t *testing.T) {
	ctx := context.Background()
	u := NewUnavailableOfferings(0)

	assert.False(t, u.IsUnavailable("VM.Standard.E5.Flex", "AD-1", "spot"))

	u.MarkUnavailable(ctx, "VM.Standard.E5.Flex", "AD-1", "spot")
	assert.True(t, u.IsUnavailable("VM.Standard.E5.Flex", "AD-1", "spot"))

	// distinct capacity type, zone and shape are unaffected.
	assert.False(t, u.IsUnavailable("VM.Standard.E5.Flex", "AD-1", "on-demand"))
	assert.False(t, u.IsUnavailable("VM.Standard.E5.Flex", "AD-2", "spot"))
	assert.False(t, u.IsUnavailable("VM.Standard.E4.Flex", "AD-1", "spot"))
}

func TestUnavailableOfferings_Flush(t *testing.T) {
	ctx := context.Background()
	u := NewUnavailableOfferings(0)

	u.MarkUnavailable(ctx, "VM.Standard.E5.Flex", "AD-1", "spot")
	u.MarkUnavailable(ctx, "VM.Standard.E5.Flex", "AD-2", "on-demand")
	assert.True(t, u.IsUnavailable("VM.Standard.E5.Flex", "AD-1", "spot"))
	assert.True(t, u.IsUnavailable("VM.Standard.E5.Flex", "AD-2", "on-demand"))

	u.Flush()
	assert.False(t, u.IsUnavailable("VM.Standard.E5.Flex", "AD-1", "spot"))
	assert.False(t, u.IsUnavailable("VM.Standard.E5.Flex", "AD-2", "on-demand"))
}

func TestUnavailableOfferings_TTLExpiry(t *testing.T) {
	ctx := context.Background()
	// a short, configurable TTL is honored by MarkUnavailable so entries expire quickly.
	u := NewUnavailableOfferings(20 * time.Millisecond)

	u.MarkUnavailable(ctx, "VM.Standard.E5.Flex", "AD-1", "spot")
	assert.True(t, u.IsUnavailable("VM.Standard.E5.Flex", "AD-1", "spot"))

	time.Sleep(40 * time.Millisecond)
	assert.False(t, u.IsUnavailable("VM.Standard.E5.Flex", "AD-1", "spot"))
}

func TestUnavailableOfferings_DefaultsTTLWhenNonPositive(t *testing.T) {
	assert.Equal(t, UnavailableOfferingsTTL, NewUnavailableOfferings(0).ttl)
	assert.Equal(t, UnavailableOfferingsTTL, NewUnavailableOfferings(-time.Second).ttl)
	assert.Equal(t, time.Minute, NewUnavailableOfferings(time.Minute).ttl)
}
