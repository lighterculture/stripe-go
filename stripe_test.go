package stripe_test

import (
	"encoding/json"
	"net/http"
	"runtime"
	"testing"

	stripe "github.com/stripe/stripe-go"
	. "github.com/stripe/stripe-go/testing"
)

func TestCheckinBackendConfigurationNewRequestWithStripeAccount(t *testing.T) {
	c := &stripe.BackendConfiguration{URL: stripe.APIURL}
	p := &stripe.Params{StripeAccount: TestMerchantID}

	req, err := c.NewRequest("", "", "", "", nil, p)
	if err != nil {
		t.Fatal(err)
	}

	if req.Header.Get("Stripe-Account") != TestMerchantID {
		t.Fatalf("Expected Stripe-Account %v but got %v.",
			TestMerchantID, req.Header.Get("Stripe-Account"))
	}

	// Also test the deprecated Account field for now as well. This should be
	// identical to the exercise above.
	p = &stripe.Params{Account: TestMerchantID}

	req, err = c.NewRequest("", "", "", "", nil, p)
	if err != nil {
		t.Fatal(err)
	}

	if req.Header.Get("Stripe-Account") != TestMerchantID {
		t.Fatalf("Expected Stripe-Account %v but got %v.",
			TestMerchantID, req.Header.Get("Stripe-Account"))
	}
}

func TestCheckinStripeClientUserAgent(t *testing.T) {
	c := &stripe.BackendConfiguration{URL: stripe.APIURL}
	p := &stripe.Params{}

	req, err := c.NewRequest("", "", "", "", nil, p)
	if err != nil {
		t.Fatal(err)
	}

	encodedUserAgent := req.Header.Get("X-Stripe-Client-User-Agent")
	if encodedUserAgent == "" {
		t.Fatalf("Expected X-Stripe-Client-User-Agent header to be present.")
	}

	var userAgent map[string]string
	err = json.Unmarshal([]byte(encodedUserAgent), &userAgent)
	if err != nil {
		t.Fatal(err)
	}

	//
	// Just test a few headers that we know to be stable.
	//

	if userAgent["language"] != "go" {
		t.Fatalf("Expected X-Stripe-Client-User-Agent/language %v but got %v.",
			"go", userAgent["language"])
	}

	if userAgent["language_version"] != runtime.Version() {
		t.Fatalf("Expected X-Stripe-Client-User-Agent/language_version %v but got %v.",
			runtime.Version(), userAgent["language_version"])
	}

	// Anywhere these tests are running can reasonable be expected to have a
	// `uname` to run, so do this basic check.
	if userAgent["uname"] == stripe.UnknownPlatform {
		t.Fatalf("Expected X-Stripe-Client-User-Agent/uname to have a value.")
	}
}

func TestCheckinResponseToError(t *testing.T) {
	c := &stripe.BackendConfiguration{URL: stripe.APIURL}

	// A test response that includes a status code and request ID.
	res := &http.Response{
		Header: http.Header{
			"Request-Id": []string{"request-id"},
		},
		StatusCode: 402,
	}

	// An error that contains expected fields which we're going to serialize to
	// JSON and inject into our converstion function.
	expectedErr := &stripe.Error{
		Code:  stripe.Missing,
		Msg:   "That card was declined",
		Param: "expiry_date",
		Type:  stripe.ErrorTypeCard,
	}
	bytes, err := json.Marshal(expectedErr)
	if err != nil {
		t.Fatal(err)
	}

	// Unpack the error that we just serialized so that we can inject a
	// type-specific field into it ("decline_code"). This will show up in a
	// field on a special stripe.CardError type which is attached to the common
	// stripe.Error.
	var raw map[string]string
	err = json.Unmarshal(bytes, &raw)
	if err != nil {
		t.Fatal(err)
	}
	expectedDeclineCode := "decline-code"
	raw["decline_code"] = expectedDeclineCode
	bytes, err = json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}

	// A generic Golang error.
	err = c.ResponseToError(res, wrapError(bytes))

	// An error containing Stripe-specific fields that we cast back from the
	// generic Golang error.
	stripeErr := err.(*stripe.Error)

	if expectedErr.Code != stripeErr.Code {
		t.Fatalf("Expected code %v but got %v.", expectedErr.Code, stripeErr.Code)
	}

	if expectedErr.Msg != stripeErr.Msg {
		t.Fatalf("Expected message %v but got %v.", expectedErr.Msg, stripeErr.Msg)
	}

	if expectedErr.Param != stripeErr.Param {
		t.Fatalf("Expected param %v but got %v.", expectedErr.Param, stripeErr.Param)
	}

	if res.Header.Get("Request-Id") != stripeErr.RequestID {
		t.Fatalf("Expected code %v but got %v.", res.Header.Get("Request-Id"), stripeErr.RequestID)
	}

	if res.StatusCode != stripeErr.HTTPStatusCode {
		t.Fatalf("Expected code %v but got %v.", res.StatusCode, stripeErr.HTTPStatusCode)
	}

	if expectedErr.Type != stripeErr.Type {
		t.Fatalf("Expected type %v but got %v.", expectedErr.Type, stripeErr.Type)
	}

	// Just a bogus type coercion to demonstrate how this code might be
	// written. Because we've assigned ErrorTypeCard as the error's type, Err
	// should always come out as a CardError.
	_, ok := stripeErr.Err.(*stripe.InvalidRequestError)
	if ok {
		t.Fatalf("Expected error to not be coercable to InvalidRequestError.")
	}

	cardErr, ok := stripeErr.Err.(*stripe.CardError)
	if !ok {
		t.Fatalf("Expected error to be coercable to CardError.")
	}

	if expectedDeclineCode != cardErr.DeclineCode {
		t.Fatalf("Expected decline code %v but got %v.", expectedDeclineCode, cardErr.DeclineCode)
	}
}

// A simple function that allows us to represent an error response from Stripe
// which comes wrapper in a JSON object with a single field of "error".
func wrapError(serialized []byte) []byte {
	return []byte(`{"error":` + string(serialized) + `}`)
}
