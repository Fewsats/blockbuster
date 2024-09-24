package lightning_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/fewsats/blockbuster/lightning"
	"github.com/stretchr/testify/require"
)

var albyRespBody = `{
	"amount":16,
	"boostagram":null,
	"comment":null,
	"created_at":"2024-05-15T17:43:44.032Z",
	"creation_date":1715795024,
	"currency":"USD",
	"custom_records":null,
	"description_hash":null,
	"expires_at":"2024-05-16T17:43:44.000Z",
	"expiry":86400,
	"fiat_currency":"USD",
	"fiat_in_cents":1,
	"identifier":"tLWnjJwaZGnwLM4YdWDmVvec",
	"keysend_message":null,
	"memo":null,"payer_name":null,
	"payer_email":null,
	"payer_pubkey":null,
	"payment_hash":"9f85cb6454c7b7524540c5c6c6b6ccb7da757be2199390c4d10b5795f1718871",
	"payment_request":"lnbc160n1pnyfazspp5n7zukez5c7m4y32qchrvddkvkld827lzrxfep3x3pdtetut33pcsdqqcqzzsxqyz5vqsp5w5wsm5dtv7jf9ft227neh6wzdvznfmyu5zyj58w6ydwz06dg0sfs9qyyssqds8q4jjf74nvcln8zl6frrrqellcqr303raa0k3scgvpexj355kj0hcfsav3048cz747hjdxkqyzykvj6drjhwxjzdamzykh6acudsgpsthcfj",
	"preimage":null,
	"r_hash_str":"9f85cb6454c7b7524540c5c6c6b6ccb7da757be2199390c4d10b5795f1718871",
	"settled":false,
	"settled_at":null,
	"state":"CREATED",
	"type":"incoming",
	"value":16,
	"metadata":null,
	"destination_alias":null,
	"destination_pubkey":null,
	"first_route_hint_pubkey":null,
	"first_route_hint_alias":null,
	"qr_code_png":"https://getalby.com/api/invoices/tLWnjJwaZGnwLM4YdWDmVvec.png",
	"qr_code_svg":"https://getalby.com/api/invoices/tLWnjJwaZGnwLM4YdWDmVvec.svg"
}`

// TestAlbyClient tests the AlbyClient.
func TestAlbyClient(t *testing.T) {
	t.Helper()

	httpClient := &MockHTTPClient{
		DoFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader(albyRespBody)),
			}, nil
		},
	}

	// Create a new AlbyClient.
	albyProvider := lightning.NewAlbyProvider(httpClient, "fakeToken")

	ctx := context.Background()
	amount := uint64(1)
	currency := "USD"
	description := ""

	invoice, err := albyProvider.CreateInvoice(
		ctx, amount, currency, description,
	)

	require.NoError(t, err)

	expectedInvoiceAmount := uint64(16)
	expectedInvoiceCurrency := "BTC"
	expectedPaymentHash := "9f85cb6454c7b7524540c5c6c6b6ccb7da757be2199390c4d10b5795f1718871"
	expectedPaymentRequest := "lnbc160n1pnyfazspp5n7zukez5c7m4y32qchrvddkvkld827lzrxfep3x3pdtetut33pcsdqqcqzzsxqyz5vqsp5w5wsm5dtv7jf9ft227neh6wzdvznfmyu5zyj58w6ydwz06dg0sfs9qyyssqds8q4jjf74nvcln8zl6frrrqellcqr303raa0k3scgvpexj355kj0hcfsav3048cz747hjdxkqyzykvj6drjhwxjzdamzykh6acudsgpsthcfj"

	require.Equal(t, amount, invoice.UserAmount.Amount)
	require.Equal(t, currency, invoice.UserAmount.Currency)
	require.Equal(t, expectedInvoiceAmount, invoice.PaymentAmount.Amount)
	require.Equal(t, expectedInvoiceCurrency, invoice.PaymentAmount.Currency)
	require.Equal(t, expectedPaymentHash, invoice.PaymentHash)
	require.Equal(t, expectedPaymentRequest, invoice.PaymentRequest)
}

func TestCreateInvoiceHandlesHTTPError(t *testing.T) {
	t.Helper()

	httpClient := &MockHTTPClient{
		DoFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized, // 401 Unauthorized
				Body:       ioutil.NopCloser(strings.NewReader("Unauthorized access")),
			}, nil
		},
	}

	albyProvider := lightning.NewAlbyProvider(httpClient, "invalidToken")

	ctx := context.Background()
	amount := uint64(100)
	currency := "USD"
	description := "Test invoice"

	_, err := albyProvider.CreateInvoice(ctx, amount, currency, description)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "401")
}
