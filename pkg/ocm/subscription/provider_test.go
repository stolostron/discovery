package subscription

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

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
func TestProviderGetSubscriptionsNoError(t *testing.T) {
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
	httpClient = &getClientMock{} //without this line, the real api is fired

	response, err := SubscriptionProvider.GetSubscriptions(SubscriptionRequest{})
	assert.NotNil(t, response)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(response.Items))
}
