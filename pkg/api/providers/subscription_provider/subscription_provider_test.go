package subscription_provider

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/open-cluster-management/discovery/pkg/api/clients/restclient"
	"github.com/open-cluster-management/discovery/pkg/api/domain/subscription_domain"

	"github.com/stretchr/testify/assert"
)

var (
	getRequestFunc func(*http.Request) (*http.Response, error)
)

// Mocking the SubscriptionGetInterface
type getClientMock struct{}

func (cm *getClientMock) Get(request *http.Request) (*http.Response, error) {
	return getRequestFunc(request)
}

//When the everything is good
func TestGetSubscriptionsNoError(t *testing.T) {
	getRequestFunc = func(*http.Request) (*http.Response, error) {
		file, err := os.Open("testdata/accounts_mgmt_mock.json")
		if err != nil {
			t.Error(err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(file),
		}, nil
	}
	restclient.SubscriptionHTTPClient = &getClientMock{} //without this line, the real api is fired

	response, err := SubscriptionProvider.GetSubscriptions(subscription_domain.SubscriptionRequest{})
	assert.NotNil(t, response)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, len(response.Items))
}
