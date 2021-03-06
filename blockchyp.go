package blockchyp

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

/*
Default client configuration constants.
*/
const (
	DefaultGatewayHost     = "https://api.blockchyp.com"
	DefaultTestGatewayHost = "https://test.blockchyp.com"
	DefaultHTTPS           = true
	DefaultRouteCacheTTL   = 60 * time.Minute
	DefaultGatewayTimeout  = 20 * time.Second
	DefaultTerminalTimeout = 2 * time.Minute
)

/*
Default filesystem configuration.
*/
const (
	ConfigDir  = "blockchyp"
	ConfigFile = "blockchyp.json"
)

// terminalCN is the common name on a terminal certificate.
const terminalCN = "blockchyp-terminal"

/*
Client is the main interface used by application developers.
*/
type Client struct {
	Credentials        APICredentials
	GatewayHost        string
	TestGatewayHost    string
	HTTPS              bool
	RouteCache         string
	routeCacheTTL      time.Duration
	gatewayHTTPClient  *http.Client
	terminalHTTPClient *http.Client
}

/*
NewClient returns a default Client configured with the given credentials.
*/
func NewClient(creds APICredentials) Client {
	return Client{
		Credentials:   creds,
		GatewayHost:   DefaultGatewayHost,
		HTTPS:         DefaultHTTPS,
		routeCacheTTL: DefaultRouteCacheTTL,
		RouteCache:    filepath.Join(os.TempDir(), ".blockchyp_routes"),
		gatewayHTTPClient: &http.Client{
			Timeout: DefaultGatewayTimeout,
		},
		terminalHTTPClient: &http.Client{
			Timeout: DefaultTerminalTimeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:    terminalCertPool(),
					ServerName: terminalCN,
				},
			},
		},
	}
}

/*
AsyncCharge executes an asynchronous auth and capture.
*/
func (client *Client) AsyncCharge(request AuthorizationRequest, responseChan chan<- AuthorizationResponse) error {

	if !isValidAsyncMethod(request.PaymentMethod) {
		return newInvalidAsyncRequestError()
	}

	return nil
}

/*
Charge executes a standard direct preauth and capture.
*/
func (client *Client) Charge(request AuthorizationRequest) (*AuthorizationResponse, error) {

	if isTerminalRouted(request.PaymentMethod) {
		route, err := client.resolveTerminalRoute(request.TerminalName)
		if err != nil {
			return nil, err
		}
		if !route.CloudRelayEnabled {
			authResponse := AuthorizationResponse{}
			if !route.Exists {
				authResponse.Approved = false
				authResponse.ResponseDescription = "Unknown Terminal"
				return &authResponse, err
			}
			authRequest := TerminalAuthorizationRequest{
				APICredentials: route.TransientCredentials,
				Request:        request,
			}
			err = client.terminalPost(route, "/charge", authRequest, &authResponse)
			if err, ok := err.(net.Error); ok && err.Timeout() {
				authResponse.Approved = false
				authResponse.ResponseDescription = "Request Timed Out"
			} else if err != nil {
				authResponse.Approved = false
				authResponse.ResponseDescription = err.Error()
			}
			return &authResponse, err
		}
	}
	authResponse := AuthorizationResponse{}
	err := client.GatewayPost("/charge", request, &authResponse, request.Test)
	return &authResponse, err

}

/*
AsyncPreauth executes an asynchronous preauthorization.
*/
func (client *Client) AsyncPreauth(request AuthorizationRequest, responseChan chan<- AuthorizationResponse) error {

	if !isValidAsyncMethod(request.PaymentMethod) {
		return newInvalidAsyncRequestError()
	}

	return nil
}

/*
Preauth executes a preauthorization intended to be captured later.
*/
func (client *Client) Preauth(request AuthorizationRequest) (*AuthorizationResponse, error) {

	if isTerminalRouted(request.PaymentMethod) {
		route, err := client.resolveTerminalRoute(request.TerminalName)
		if err != nil {
			return nil, err
		}
		if !route.CloudRelayEnabled {
			authResponse := AuthorizationResponse{}
			if !route.Exists {
				authResponse.Approved = false
				authResponse.ResponseDescription = "Unknown Terminal"
				return &authResponse, err
			}
			authRequest := TerminalAuthorizationRequest{
				APICredentials: route.TransientCredentials,
				Request:        request,
			}
			err = client.terminalPost(route, "/preauth", authRequest, &authResponse)
			if err, ok := err.(net.Error); ok && err.Timeout() {
				authResponse.Approved = false
				authResponse.ResponseDescription = "Request Timed Out"
			} else if err != nil {
				authResponse.Approved = false
				authResponse.ResponseDescription = err.Error()
			}
			return &authResponse, err
		}
	}

	authResponse := AuthorizationResponse{}
	err := client.GatewayPost("/preauth", request, &authResponse, request.Test)
	return &authResponse, err

}

/*
AsyncRefund executes an asynchronous refund
*/
func (client *Client) AsyncRefund(request AuthorizationRequest, responseChan chan<- AuthorizationResponse) error {

	if !isValidAsyncMethod(request.PaymentMethod) {
		return newInvalidAsyncRequestError()
	}

	return nil
}

/*
Refund executes a refund.
*/
func (client *Client) Refund(request RefundRequest) (*AuthorizationResponse, error) {

	if isTerminalRouted(request.PaymentMethod) {
		route, err := client.resolveTerminalRoute(request.TerminalName)
		if err != nil {
			return nil, err
		}
		if !route.CloudRelayEnabled {
			authResponse := AuthorizationResponse{}
			if !route.Exists {
				authResponse.Approved = false
				authResponse.ResponseDescription = "Unknown Terminal"
				return &authResponse, err
			}
			authRequest := TerminalRefundAuthorizationRequest{
				APICredentials: route.TransientCredentials,
				Request:        request,
			}
			err = client.terminalPost(route, "/refund", authRequest, &authResponse)
			if err, ok := err.(net.Error); ok && err.Timeout() {
				authResponse.Approved = false
				authResponse.ResponseDescription = "Request Timed Out"
			} else if err != nil {
				authResponse.Approved = false
				authResponse.ResponseDescription = err.Error()
			}
			return &authResponse, err
		}
	}
	authResponse := AuthorizationResponse{}
	err := client.GatewayPost("/refund", request, &authResponse, request.Test)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		authResponse.Approved = false
		authResponse.ResponseDescription = "Request Timed Out"
	} else if err != nil {
		authResponse.Approved = false
		authResponse.ResponseDescription = err.Error()
	}
	return &authResponse, err

}

/*
Reverse executes a manual time out reversal.
*/
func (client *Client) Reverse(request AuthorizationRequest) (*AuthorizationResponse, error) {

	authResponse := AuthorizationResponse{}
	err := client.GatewayPost("/reverse", request, &authResponse, request.Test)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		authResponse.Approved = false
		authResponse.ResponseDescription = "Request Timed Out"
	} else if err != nil {
		authResponse.Approved = false
		authResponse.ResponseDescription = err.Error()
	}
	return &authResponse, err

}

/*
Capture captures a preauthorization.
*/
func (client *Client) Capture(request CaptureRequest) (*CaptureResponse, error) {

	captureResponse := CaptureResponse{}
	err := client.GatewayPost("/capture", request, &captureResponse, request.Test)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		captureResponse.Approved = false
		captureResponse.ResponseDescription = "Request Timed Out"
	} else if err != nil {
		captureResponse.Approved = false
		captureResponse.ResponseDescription = err.Error()
	}
	return &captureResponse, err

}

/*
Void discards a previous preauth transaction.
*/
func (client *Client) Void(request VoidRequest) (*VoidResponse, error) {

	voidResponse := VoidResponse{}
	err := client.GatewayPost("/void", request, &voidResponse, request.Test)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		voidResponse.Approved = false
		voidResponse.ResponseDescription = "Request Timed Out"
	} else if err != nil {
		voidResponse.Approved = false
		voidResponse.ResponseDescription = err.Error()
	}
	return &voidResponse, err
}

/*
AsyncEnroll executes an asynchronous vault enrollment.
*/
func (client *Client) AsyncEnroll(request EnrollRequest, responseChan chan<- EnrollResponse) error {

	if !isValidAsyncMethod(request.PaymentMethod) {
		return newInvalidAsyncRequestError()
	}

	return nil
}

/*
Enroll adds a new payment method to the token vault.
*/
func (client *Client) Enroll(request EnrollRequest) (*EnrollResponse, error) {

	if isTerminalRouted(request.PaymentMethod) {
		_, err := client.resolveTerminalRoute(request.TerminalName)
		if err != nil {
			return nil, err
		}

	} else {
		enrollResponse := EnrollResponse{}
		err := client.GatewayPost("/enroll", request, &enrollResponse, request.Test)
		return &enrollResponse, err
	}

	return &EnrollResponse{}, nil
}

/*
Ping tests connectivity with a payment terminal.
*/
func (client *Client) Ping(request PingRequest) (*PingResponse, error) {
	route, err := client.resolveTerminalRoute(request.TerminalName)
	if err != nil {
		return nil, err
	}

	var pingResponse PingResponse

	if !route.Exists {
		pingResponse.Success = false
		pingResponse.ResponseDescription = "Unknown Terminal"
		return &pingResponse, err
	}

	terminalRequest := TerminalPingRequest{
		APICredentials: route.TransientCredentials,
		Request:        request,
	}

	if route.CloudRelayEnabled {
		err = client.GatewayPost("/terminal-test", request, &pingResponse, request.Test)
	} else {
		err = client.terminalPost(route, "/test", terminalRequest, &pingResponse)
	}

	if err, ok := err.(net.Error); ok && err.Timeout() {
		pingResponse.Success = false
		pingResponse.ResponseDescription = "Request Timed Out"
	} else if err != nil {
		pingResponse.Success = false
		pingResponse.ResponseDescription = err.Error()
	}

	return &pingResponse, err
}

/*
GiftActivate activates or recharges a gift card.
*/
func (client *Client) GiftActivate(request GiftActivateRequest) (*GiftActivateResponse, error) {
	route, err := client.resolveTerminalRoute(request.TerminalName)
	if err != nil {
		return nil, err
	}
	terminalRequest := TerminalGiftActivateRequest{
		APICredentials: route.TransientCredentials,
		Request:        request,
	}
	giftResponse := GiftActivateResponse{}
	err = client.terminalPost(route, "/gift-activate", terminalRequest, &giftResponse)
	return &giftResponse, err
}

/*
CloseBatch closes the current credit card batch.
*/
func (client *Client) CloseBatch(request CloseBatchRequest) (*CloseBatchResponse, error) {

	response := CloseBatchResponse{}
	err := client.GatewayPost("/close-batch", request, &response, request.Test)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		response.Success = false
		response.ResponseDescription = "Request Timed Out"
	} else if err != nil {
		response.Success = false
		response.ResponseDescription = err.Error()
	}
	return &response, err

}

// NewTransactionDisplay displays a new transaction on the terminal.
func (client *Client) NewTransactionDisplay(request TransactionDisplayRequest) error {
	return client.sendTransactionDisplay(request, http.MethodPost)
}

// UpdateTransactionDisplay appends items to an existing transaction display.
// Subtotal, Tax, and Total are overwritten by the request. Items with the same
// description are combined into groups.
func (client *Client) UpdateTransactionDisplay(request TransactionDisplayRequest) error {
	return client.sendTransactionDisplay(request, http.MethodPut)
}

// ClearTransactionDisplay resets the displayed transaction and returns the
// terminal to idle.
func (client *Client) ClearTransactionDisplay(terminalName string) error {
	request := TransactionDisplayRequest{
		TerminalName: terminalName,
	}

	return client.sendTransactionDisplay(request, http.MethodDelete)
}

// sendTransactionDisplay sends a transaction display request to a terminal.
func (client *Client) sendTransactionDisplay(request TransactionDisplayRequest, method string) error {
	route, err := client.resolveTerminalRoute(request.TerminalName)
	if err != nil {
		return err
	}

	if !route.Exists {
		return errors.New("Unknown Terminal")
	}

	response := &Acknowledgement{}
	if route.CloudRelayEnabled {
		err = client.GatewayRequest("/terminal-txdisplay", method, request, response, false)
	} else {
		terminalRequest := TerminalTransactionDisplayRequest{
			APICredentials: route.TransientCredentials,
			Request:        request,
		}
		err = client.terminalRequest(route, "/txdisplay", method, terminalRequest, response)
	}
	if err != nil {
		return err
	}

	if !response.Success {
		return errors.New(response.Error)
	}

	return nil
}

func isValidAsyncMethod(method PaymentMethod) bool {

	if method.TerminalName == "" {
		return false
	} else if method.Token != "" {
		return false
	} else if method.Track1 != "" {
		return false
	} else if method.Track2 != "" {
		return false
	} else if method.PAN != "" {
		return false
	}

	return true

}

func newInvalidAsyncRequestError() error {
	return errors.New("async requests must be terminal requests")
}
