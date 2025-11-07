package service

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	istioasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/istio"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

//go:embed service_fallback.yaml
var APIRuleServiceFallback string

//go:embed service_two_namespaces.yaml
var APIRuleServiceTwoNamespaces string

//go:embed service_diff_same_methods.yaml
var APIRuleServiceDiffSameMethods string

//go:embed service_custom_label_selector.yaml
var APIRuleServiceCustomLabelSelector string

func TestAPIRuleValidation(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

	t.Run("Endpoints exposed in APIRule should fallback to service defined on root level when there is no service defined on rule level", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("no-auth"))
		require.NoError(t, err, "Failed to setup test background with httpbin")
		kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
		require.NoError(t, err, "Failed to get domain from kyma-gateway")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleServiceFallback,
			map[string]any{
				"Name":                         testBackground.TestName,
				"Host":                         testBackground.TestName,
				"ServiceName":                  testBackground.TargetServiceName,
				"ServicePort":                  testBackground.TargetServicePort,
				"Gateway":                      "kyma-system/kyma-gateway",
				"Issuer":                       testBackground.Provider.GetIssuerURL(),
				"JwksUri":                      testBackground.Provider.GetJwksURI(),
				"JwtSecuredPathWithService":    "/headers",
				"JwtSecuredPathWithoutService": "/ip",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)

		urlWithServiceDefined := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/headers")
		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			urlWithServiceDefined,
			http.MethodGet,
		)

		urlWithoutServiceDefined := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/ip")
		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			urlWithoutServiceDefined,
			http.MethodGet,
		)
	})

	//t.Run("Exposing endpoints in two namespaces", func(t *testing.T) {
	//	testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("no-auth"))
	//	require.NoError(t, err, "Failed to setup test background with httpbin")
	//
	//	createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
	//		t,
	//		APIRuleServiceTwoNamespaces,
	//		// got to fulfill these properly
	//		map[string]any{
	//			"Name":        testBackground.TestName,
	//			"Host":        testBackground.TestName,
	//			"ServiceName": testBackground.TargetServiceName,
	//			"ServicePort": testBackground.TargetServicePort,
	//			"Gateway":     "kyma-system/kyma-gateway",
	//		},
	//		decoder.MutateNamespace(testBackground.Namespace),
	//	)
	//	require.NoError(t, err, "Failed to create APIRule resource")
	//	require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")
	//
	//	apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
	//	istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
	//	istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)
	//
	//	requests := []struct {
	//		path                   string
	//		method                 string
	//		expectedResponseStatus int
	//	}{
	//		{path: "/status/200", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
	//		{path: "/headers", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
	//	}
	//
	//	for _, r := range requests {
	//		println(r.path)
	//	}
	//})
	//
	//t.Run("Exposing different services with same methods", func(t *testing.T) {
	//	testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("no-auth"))
	//	require.NoError(t, err, "Failed to setup test background with httpbin")
	//	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	//	require.NoError(t, err, "Failed to get domain from kyma-gateway")
	//
	//	createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
	//		t,
	//		APIRuleServiceDiffSameMethods,
	//		// got to fulfill these properly
	//		map[string]any{
	//			"Name":        testBackground.TestName,
	//			"Host":        testBackground.TestName,
	//			"ServiceName": testBackground.TargetServiceName,
	//			"ServicePort": testBackground.TargetServicePort,
	//			"Gateway":     "kyma-system/kyma-gateway",
	//		},
	//		decoder.MutateNamespace(testBackground.Namespace),
	//	)
	//	require.NoError(t, err, "Failed to create APIRule resource")
	//	require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")
	//
	//	apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
	//	apiruleasserts.HasAnnotation(t, testBackground.TestName, testBackground.Namespace, "gateway.kyma-project.io/original-version", "v2")
	//	istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
	//	istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)
	//
	//	requests := []struct {
	//		path                   string
	//		method                 string
	//		expectedResponseStatus int
	//	}{
	//		{path: "/ip", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
	//		{path: "/status/200", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
	//		{path: "/headers", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
	//	}
	//
	//	for _, r := range requests {
	//		url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, r.path)
	//		req, err := http.NewRequest(r.method, url, nil)
	//		if err != nil {
	//			t.Fatalf("err %s", err.Error())
	//		}
	//		c := h.NewHTTPClient(t)
	//		resp, err := c.Do(req)
	//		if err != nil {
	//			t.Fatalf("err %s", err.Error())
	//		}
	//		require.Equal(t, resp.StatusCode, r.expectedResponseStatus)
	//	}
	//})
	//
	//t.Run("Calling a helloworld endpoint with custom label selector service", func(t *testing.T) {
	//	testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("no-auth"))
	//	require.NoError(t, err, "Failed to setup test background with httpbin")
	//	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	//	require.NoError(t, err, "Failed to get domain from kyma-gateway")
	//
	//	createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
	//		t,
	//		APIRuleServiceCustomLabelSelector,
	//		// got to fulfill these properly
	//		map[string]any{
	//			"Name":        testBackground.TestName,
	//			"Host":        testBackground.TestName,
	//			"ServiceName": testBackground.TargetServiceName,
	//			"ServicePort": testBackground.TargetServicePort,
	//			"Gateway":     "kyma-system/kyma-gateway",
	//		},
	//		decoder.MutateNamespace(testBackground.Namespace),
	//	)
	//	require.NoError(t, err, "Failed to create APIRule resource")
	//	require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")
	//
	//	apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
	//	apiruleasserts.HasAnnotation(t, testBackground.TestName, testBackground.Namespace, "gateway.kyma-project.io/original-version", "v2")
	//	istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
	//	istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)
	//
	//	requests := []struct {
	//		path                   string
	//		method                 string
	//		expectedResponseStatus int
	//	}{
	//		{path: "/ip", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
	//		{path: "/status/200", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
	//		{path: "/headers", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
	//	}
	//
	//	for _, r := range requests {
	//		url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, r.path)
	//		req, err := http.NewRequest(r.method, url, nil)
	//		if err != nil {
	//			t.Fatalf("err %s", err.Error())
	//		}
	//		c := h.NewHTTPClient(t)
	//		resp, err := c.Do(req)
	//		if err != nil {
	//			t.Fatalf("err %s", err.Error())
	//		}
	//		require.Equal(t, resp.StatusCode, r.expectedResponseStatus)
	//	}
	//})
}
