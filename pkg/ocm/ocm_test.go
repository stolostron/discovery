package ocm

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	discovery "github.com/open-cluster-management/discovery/api/v1alpha1"
	"github.com/open-cluster-management/discovery/pkg/ocm/auth"
	"github.com/open-cluster-management/discovery/pkg/ocm/subscription"
)

var (
	getTokenFunc func(auth.AuthRequest) (string, error)

	getSubscriptionsFunc func() ([]subscription.Subscription, error)
	subscriptionGetter   = subscriptionGetterMock{}
)

// This mocks the authService request and returns a dummy access token
type authServiceMock struct{}

func (m *authServiceMock) GetToken(request auth.AuthRequest) (string, error) {
	return getTokenFunc(request)
}

// The mocks the GetClusters request to return a select few clusters without connection
// to an external datasource
type subscriptionGetterMock struct{}

func (m *subscriptionGetterMock) GetSubscriptions() ([]subscription.Subscription, error) {
	return getSubscriptionsFunc()
}

// This mocks the NewClient function and returns an instance of the subscriptionGetterMock
type subscriptionClientGeneratorMock struct{}

func (m *subscriptionClientGeneratorMock) NewClient(config subscription.SubscriptionRequest) subscription.SubscriptionGetter {
	return &subscriptionGetter
}

// clustersResponse takes in a file with subscription data and returns a new mock function
func subscriptionResponse(testdata string) func() ([]subscription.Subscription, error) {
	return func() ([]subscription.Subscription, error) {
		file, _ := ioutil.ReadFile(testdata)
		subscriptions := []subscription.Subscription{}
		err := json.Unmarshal([]byte(file), &subscriptions)
		return subscriptions, err
	}
}

func TestDiscoverClusters(t *testing.T) {
	type args struct {
		token   string
		baseURL string
		filters discovery.Filter
	}
	tests := []struct {
		name             string
		authfunc         func(auth.AuthRequest) (string, error)
		subscriptionFunc func() ([]subscription.Subscription, error)
		args             args
		want             int
		wantErr          bool
	}{
		{
			name: "Complete subscription",
			authfunc: func(auth.AuthRequest) (string, error) {
				// this mock return a dummy token
				return "valid_access_token", nil
			},
			// this mock returns 3 subscriptions read from mock_subscriptions.json
			subscriptionFunc: subscriptionResponse("testdata/1_mock_subscription.json"),
			args: args{
				token:   "test",
				baseURL: "test",
				filters: discovery.Filter{},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Two complete subscriptions, one incomplete",
			authfunc: func(auth.AuthRequest) (string, error) {
				// this mock return a dummy token
				return "valid_access_token", nil
			},
			// this mock returns 3 subscriptions read from mock_subscriptions.json
			subscriptionFunc: subscriptionResponse("testdata/3_mock_subscriptions.json"),
			args: args{
				token:   "test",
				baseURL: "test",
				filters: discovery.Filter{},
			},
			want:    2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth.AuthClient = &authServiceMock{}                                          // Mocks out the call to auth service
			subscription.SubscriptionClientGenerator = &subscriptionClientGeneratorMock{} // Mocks out the subscription client creation

			getTokenFunc = tt.authfunc
			getSubscriptionsFunc = tt.subscriptionFunc

			got, err := DiscoverClusters(tt.args.token, tt.args.baseURL, tt.args.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiscoverClusters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("DiscoverClusters() = %v, wanted %d clusters", got, tt.want)
			}
		})
	}
}

func Test_computeDisplayName(t *testing.T) {
	tests := []struct {
		name string
		sub  subscription.Subscription
		want string
	}{
		{
			name: "Custom displayname set",
			sub: subscription.Subscription{
				ConsoleURL:        "https://console-openshift-console.apps.installer-pool-j88kj.dev01.red-chesterfield.com",
				ExternalClusterID: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
				DisplayName:       "my-custom-name",
			},
			want: "my-custom-name",
		},
		{
			name: "No custom displayname - use consoleURL",
			sub: subscription.Subscription{
				ConsoleURL:        "https://console-openshift-console.apps.installer-pool-j88kj.dev01.red-chesterfield.com",
				ExternalClusterID: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
				DisplayName:       "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
			},
			want: "installer-pool-j88kj-dev01-red-chesterfield-com",
		},
		{
			name: "Displayname missing - use consoleURL",
			sub: subscription.Subscription{
				ConsoleURL:        "https://console-openshift-console.apps.installer-pool-j88kj.dev01.red-chesterfield.com",
				ExternalClusterID: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
				DisplayName:       "",
			},
			want: "installer-pool-j88kj-dev01-red-chesterfield-com",
		},
		{
			name: "Displayname and consoleURL missing - use GUID",
			sub: subscription.Subscription{
				ConsoleURL:        "",
				ExternalClusterID: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
				DisplayName:       "",
			},
			want: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
		},
		{
			name: "ConsoleURL malformed - use GUID",
			sub: subscription.Subscription{
				ConsoleURL:        "www.installer-pool-j88kj.dev01.red-chesterfield.com",
				ExternalClusterID: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
				DisplayName:       "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
			},
			want: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
		},
		{
			name: "Port in consoleURL - remove port",
			sub: subscription.Subscription{
				ConsoleURL:        "https://console-openshift-console.apps.installer-pool-j88kj.dev01.red-chesterfield.com:6443",
				ExternalClusterID: "9cf50ab1-1f8a-4205-8a84-6958d49b469b",
				DisplayName:       "",
			},
			want: "installer-pool-j88kj-dev01-red-chesterfield-com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := computeDisplayName(tt.sub); got != tt.want {
				t.Errorf("computeDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_computeApiUrl(t *testing.T) {
	tests := []struct {
		name string
		sub  subscription.Subscription
		want string
	}{
		{
			name: "Regular consoleURL",
			sub: subscription.Subscription{
				ConsoleURL: "https://console-openshift-console.apps.installer-pool-j88kj.dev01.red-chesterfield.com",
			},
			want: "https://api.installer-pool-j88kj.dev01.red-chesterfield.com:6443",
		},
		{
			name: "Irregular consoleURL",
			sub: subscription.Subscription{
				ConsoleURL: "https://console.apps.ocp.mylab.int",
			},
			want: "",
		},
		{
			name: "No consoleURL",
			sub: subscription.Subscription{
				ConsoleURL: "",
			},
			want: "",
		},
		{
			name: "Port in consoleURL",
			sub: subscription.Subscription{
				ConsoleURL: "https://console-openshift-console.apps.installer-pool-j88kj.dev01.red-chesterfield.com:443",
			},
			want: "https://api.installer-pool-j88kj.dev01.red-chesterfield.com:6443",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := computeApiUrl(tt.sub); got != tt.want {
				t.Errorf("computeApiUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
