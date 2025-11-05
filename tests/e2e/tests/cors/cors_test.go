package cors

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

//go:embed cors_default.yaml
var APIRuleCorsDefault string

//go:embed cors_custom.yaml
var APIRuleCorsCustom string

func TestAPIRuleCors(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	t.Run("No CORS headers are returned when CORS is not specified in the APIRule", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("asterisk"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleCorsDefault,
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

		url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/ip")
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Fatalf("err %s", err.Error())
		}
		c := h.NewHTTPClient(t)
		resp, err := c.Do(req)
		if err != nil {
			t.Fatalf("err %s", err.Error())
		}
		require.Equal(t, resp.StatusCode, http.StatusOK)
		require.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
		require.Empty(t, resp.Header.Get("Access-Control-Allow-Methods"))
		require.Empty(t, resp.Header.Get("Access-Control-Allow-Headers"))
		require.Empty(t, resp.Header.Get("Access-Control-Expose-Headers"))
		require.Empty(t, resp.Header.Get("Access-Control-Allow-Credentials"))
		require.Empty(t, resp.Header.Get("Access-Control-Max-Age"))

	})

	t.Run("CORS headers are returned when CORS is specified in the APIRule", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("asterisk"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleCorsCustom,
			map[string]any{
				"Name":             testBackground.TestName,
				"Host":             testBackground.TestName,
				"ServiceName":      testBackground.TargetServiceName,
				"ServicePort":      testBackground.TargetServicePort,
				"Gateway":          "kyma-system/kyma-gateway",
				"Regex":            ".*local.kyma.dev",
				"AllowedMethods":   []string{"GET", "POST"},
				"AllowedHeaders":   []string{"x-custom-allow-headers"},
				"AllowCredentials": "false",
				"ExposeHeaders":    []string{"x-custom-expose-headers"},
				"MaxAge":           "300",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)

		originHeaderValues := []string{
			"test.local.kyma.dev",
			"localhost",
			"a.local.kyma.dev",
			"b.local.kyma.dev",
			"c.local.kyma.dev",
			"d.local.kyma.dev",
		}

		for _, originHeaderValue := range originHeaderValues {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/ip")
			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.Header.Add("Origin", originHeaderValue)
			if err != nil {
				t.Fatalf("err %s", err.Error())
			}
			c := h.NewHTTPClient(t)
			resp, err := c.Do(req)
			if err != nil {
				t.Fatalf("err %s", err.Error())
			}
			if originHeaderValue == "localhost" {
				require.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
			} else {
				require.Equal(t, resp.Header.Get("Access-Control-Allow-Origin"), originHeaderValue)
			}
			// make sure these are correct once again
			require.Equal(t, resp.StatusCode, http.StatusOK)
			require.Empty(t, resp.Header.Get("Access-Control-Allow-Methods"), "GET, POST")
			require.Equal(t, resp.Header.Get("Access-Control-Allow-Headers"), "x-custom-allow-headers")
			require.Equal(t, resp.Header.Get("Access-Control-Expose-Headers"), "x-custom-expose-headers")
			require.Equal(t, resp.Header.Get("Access-Control-Max-Age"), "300")
		}
	})
}
