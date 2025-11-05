package no_auth

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	istioasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/istio"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	h "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

//go:embed no_auth_wildcard.yaml
var APIRuleNoAuthWildcard string

//go:embed no_auth_wildcard_updated.yaml
var APIRuleNoAuthWildcardUpdated string

func TestAPIRuleValidation(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

	t.Run("Calling an endpoint unsecured on all paths from outside of the cluster", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("no-auth"))
		require.NoError(t, err, "Failed to setup test background with httpbin")
		kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
		require.NoError(t, err, "Failed to get domain from kyma-gateway")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleNoAuthWildcard,
			// got to fulfill these properly
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)

		requests := []struct {
			path                   string
			method                 string
			expectedResponseStatus int
		}{
			{path: "/ip", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/status/200", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/headers", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
		}

		for _, r := range requests {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, r.path)
			req, err := http.NewRequest(r.method, url, nil)
			if err != nil {
				t.Fatalf("err %s", err.Error())
			}
			c := h.NewHTTPClient(t)
			resp, err := c.Do(req)
			if err != nil {
				t.Fatalf("err %s", err.Error())
			}
			require.Equal(t, resp.StatusCode, r.expectedResponseStatus)
		}
	})

	t.Run("Calling an endpoint unsecured on all paths from inside of the cluster", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("no-auth"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleNoAuthWildcard,
			// got to fulfill these properly
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)

		requests := []struct {
			path                   string
			method                 string
			expectedResponseStatus int
		}{
			{path: "/status/200", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/headers", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
		}

		for _, r := range requests {
			println(r.path)
		}
	})

	t.Run("Calling an endpoint unsecured on all paths from outside of the cluster", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("no-auth"))
		require.NoError(t, err, "Failed to setup test background with httpbin")
		kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
		require.NoError(t, err, "Failed to get domain from kyma-gateway")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleNoAuthWildcard,
			// got to fulfill these properly
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		apiruleasserts.HasAnnotation(t, testBackground.TestName, testBackground.Namespace, "gateway.kyma-project.io/original-version", "v2")
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)

		updatedApirule, err := infrahelpers.UpdateResourceWithTemplateValues(
			t,
			APIRuleNoAuthWildcard,
			// got to fulfill these properly
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
			})

		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, updatedApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		apiruleasserts.HasAnnotation(t, testBackground.TestName, testBackground.Namespace, "gateway.kyma-project.io/original-version", "v2")
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)

		requests := []struct {
			path                   string
			method                 string
			expectedResponseStatus int
		}{
			{path: "/ip", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/status/200", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/headers", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
		}

		for _, r := range requests {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, r.path)
			req, err := http.NewRequest(r.method, url, nil)
			if err != nil {
				t.Fatalf("err %s", err.Error())
			}
			c := h.NewHTTPClient(t)
			resp, err := c.Do(req)
			if err != nil {
				t.Fatalf("err %s", err.Error())
			}
			require.Equal(t, resp.StatusCode, r.expectedResponseStatus)
		}
	})
}
