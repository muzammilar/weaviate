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

package namespace

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/weaviate/weaviate/client/schema"
	"github.com/weaviate/weaviate/client/users"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/test/docker"
	"github.com/weaviate/weaviate/test/helper"
)

// TestNamespaces_CollectionAndAliasCreate exercises the inline qualification
// added to AddClass / AddAlias. RBAC is off so namespaced DB users reach the
// handler unconditionally; the only gating in play is the handler-level
// IsGlobalOperator/Namespace check plus the entity-name validators.
func TestNamespaces_CollectionAndAliasCreate(t *testing.T) {
	const (
		testAdminUser = "admin-user"
		testAdminKey  = "admin-key"
		ns1           = "customer1"
		ns2           = "customer2"
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	compose, err := docker.New().WithWeaviate().
		WithApiKey().WithUserApiKey(testAdminUser, testAdminKey).
		WithDbUsers().
		WithNamespaces().
		Start(ctx)
	require.NoError(t, err)
	helper.SetupClient(compose.GetWeaviate().URI())
	defer func() {
		helper.ResetClient()
		require.NoError(t, compose.Terminate(ctx))
		cancel()
	}()

	helper.CreateNamespace(t, ns1, testAdminKey)
	helper.CreateNamespace(t, ns2, testAdminKey)
	defer helper.DeleteNamespace(t, ns1, testAdminKey)
	defer helper.DeleteNamespace(t, ns2, testAdminKey)

	user1Key := createNamespacedUser(t, "u1", ns1, testAdminKey)
	user2Key := createNamespacedUser(t, "u2", ns2, testAdminKey)
	t.Cleanup(func() {
		helper.DeleteUser(t, ns1+":u1", testAdminKey)
		helper.DeleteUser(t, ns2+":u2", testAdminKey)
	})

	t.Run("each namespaced user creates Movies; raw view shows qualified names", func(t *testing.T) {
		defer helper.DeleteClassAuth(t, "customer1:Movies", testAdminKey)
		defer helper.DeleteClassAuth(t, "customer2:Movies", testAdminKey)

		createMovies(t, user1Key)
		createMovies(t, user2Key)

		// Admin (global) sees both qualified collections.
		got1 := helper.GetClassAuth(t, "customer1:Movies", testAdminKey)
		require.Equal(t, "customer1:Movies", got1.Class)
		got2 := helper.GetClassAuth(t, "customer2:Movies", testAdminKey)
		require.Equal(t, "customer2:Movies", got2.Class)
	})

	t.Run("global admin rejected with 403 on NS-enabled cluster", func(t *testing.T) {
		class := &models.Class{Class: "Movies"}
		params := schema.NewSchemaObjectsCreateParams().WithObjectClass(class)
		_, err := helper.Client(t).Schema.SchemaObjectsCreate(params, helper.CreateAuth(testAdminKey))
		require.Error(t, err)
		var forbidden *schema.SchemaObjectsCreateForbidden
		require.True(t, errors.As(err, &forbidden), "expected SchemaObjectsCreateForbidden, got %T: %v", err, err)
	})

	t.Run("namespaced caller submitting class name with ':' rejected", func(t *testing.T) {
		class := &models.Class{Class: "Customer2:Movies"}
		params := schema.NewSchemaObjectsCreateParams().WithObjectClass(class)
		_, err := helper.Client(t).Schema.SchemaObjectsCreate(params, helper.CreateAuth(user1Key))
		require.Error(t, err)
		var unproc *schema.SchemaObjectsCreateUnprocessableEntity
		require.True(t, errors.As(err, &unproc), "expected SchemaObjectsCreateUnprocessableEntity, got %T: %v", err, err)
		assert.Contains(t, unproc.Payload.Error[0].Message, "is not a valid class name")
	})

	t.Run("namespace-local alias create succeeds and is independent across namespaces", func(t *testing.T) {
		createMovies(t, user1Key)
		createMovies(t, user2Key)
		defer helper.DeleteClassAuth(t, "customer1:Movies", testAdminKey)
		defer helper.DeleteClassAuth(t, "customer2:Movies", testAdminKey)

		// user1: Films -> Movies. The handler qualifies both to customer1:.
		alias1 := &models.Alias{Alias: "Films", Class: "Movies"}
		_, err := helper.Client(t).Schema.AliasesCreate(
			schema.NewAliasesCreateParams().WithBody(alias1),
			helper.CreateAuth(user1Key),
		)
		require.NoError(t, err)

		// user2 can independently create the same short alias name.
		alias2 := &models.Alias{Alias: "Films", Class: "Movies"}
		_, err = helper.Client(t).Schema.AliasesCreate(
			schema.NewAliasesCreateParams().WithBody(alias2),
			helper.CreateAuth(user2Key),
		)
		require.NoError(t, err)

		// Admin can see both qualified aliases exist.
		got1 := helper.GetAliasWithAuthz(t, "customer1:Films", helper.CreateAuth(testAdminKey))
		require.Equal(t, "customer1:Films", got1.Alias)
		require.Equal(t, "customer1:Movies", got1.Class)
		got2 := helper.GetAliasWithAuthz(t, "customer2:Films", helper.CreateAuth(testAdminKey))
		require.Equal(t, "customer2:Films", got2.Alias)
		require.Equal(t, "customer2:Movies", got2.Class)

		// Cleanup is best-effort; the compose teardown also wipes state.
		_, _ = helper.Client(t).Schema.AliasesDelete(
			schema.NewAliasesDeleteParams().WithAliasName("customer1:Films"),
			helper.CreateAuth(testAdminKey),
		)
		_, _ = helper.Client(t).Schema.AliasesDelete(
			schema.NewAliasesDeleteParams().WithAliasName("customer2:Films"),
			helper.CreateAuth(testAdminKey),
		)
	})

	t.Run("namespaced caller submitting alias with ':' in target rejected", func(t *testing.T) {
		alias := &models.Alias{Alias: "Films", Class: "Customer2:Movies"}
		_, err := helper.Client(t).Schema.AliasesCreate(
			schema.NewAliasesCreateParams().WithBody(alias),
			helper.CreateAuth(user1Key),
		)
		require.Error(t, err)
		var unproc *schema.AliasesCreateUnprocessableEntity
		require.True(t, errors.As(err, &unproc), "expected AliasesCreateUnprocessableEntity, got %T: %v", err, err)
	})

	t.Run("namespaced caller submitting alias with ':' in alias name rejected", func(t *testing.T) {
		alias := &models.Alias{Alias: "Customer2:Films", Class: "Movies"}
		_, err := helper.Client(t).Schema.AliasesCreate(
			schema.NewAliasesCreateParams().WithBody(alias),
			helper.CreateAuth(user1Key),
		)
		require.Error(t, err)
		var unproc *schema.AliasesCreateUnprocessableEntity
		require.True(t, errors.As(err, &unproc), "expected AliasesCreateUnprocessableEntity, got %T: %v", err, err)
	})

	t.Run("global admin rejected on alias create with NS-enabled", func(t *testing.T) {
		alias := &models.Alias{Alias: "Films", Class: "Movies"}
		_, err := helper.Client(t).Schema.AliasesCreate(
			schema.NewAliasesCreateParams().WithBody(alias),
			helper.CreateAuth(testAdminKey),
		)
		require.Error(t, err)
		var forbidden *schema.AliasesCreateForbidden
		require.True(t, errors.As(err, &forbidden), "expected AliasesCreateForbidden, got %T: %v", err, err)
	})
}

func createMovies(t *testing.T, key string) {
	t.Helper()
	class := &models.Class{Class: "Movies", Vectorizer: "none"}
	params := schema.NewSchemaObjectsCreateParams().WithObjectClass(class)
	_, err := helper.Client(t).Schema.SchemaObjectsCreate(params, helper.CreateAuth(key))
	require.NoError(t, err)
}

func createNamespacedUser(t *testing.T, userID, ns, adminKey string) string {
	t.Helper()
	resp, err := helper.Client(t).Users.CreateUser(
		users.NewCreateUserParams().WithUserID(userID).WithBody(users.CreateUserBody{Namespace: ns}),
		helper.CreateAuth(adminKey),
	)
	require.NoError(t, err)
	require.NotNil(t, resp.Payload.Apikey)
	return *resp.Payload.Apikey
}
