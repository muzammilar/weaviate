//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2026 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package namespacing

import (
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/schema"
)

// SchemaManager is a single-method interface exposing alias resolution.
// It allows the resolver to look up aliases without depending on the full
// schema reader.
type SchemaManager interface {
	ResolveAlias(alias string) string
}

// qualify prepends principal.Namespace to name.
func qualify(principal *models.Principal, name string) string {
	if principal == nil || principal.Namespace == "" {
		return name
	}
	return principal.Namespace + schema.NamespaceSeparator + name
}

// Resolve is the read-side entry point used everywhere a user-supplied
// class/alias name needs to become an internal class name:
//
//  1. Qualify the input with the principal's namespace.
//  2. Look the qualified name up as an alias via the existing in-memory
//     resolver; if it matches an alias, return the alias target.
//
// Returns (class, originalAlias, err). originalAlias is the caller-supplied
// short name when an alias was hit, "" otherwise — used by the objects
// layer to preserve existing alias-aware flows. Sites that do not need
// this can ignore the second return value.
func Resolve(principal *models.Principal, sm SchemaManager, name string) (class, originalAlias string, err error) {
	qualified := qualify(principal, name)

	// Check if the qualified name is an alias
	if resolvedClass := sm.ResolveAlias(qualified); resolvedClass != "" {
		return resolvedClass, qualified, nil
	}

	// Not an alias, return the qualified name
	return qualified, "", nil
}
