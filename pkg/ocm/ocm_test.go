package ocm

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	discoveryv1 "github.com/open-cluster-management/discovery/api/v1"
	"github.com/open-cluster-management/discovery/pkg/ocm/domain/auth_domain"
	"github.com/open-cluster-management/discovery/pkg/ocm/domain/cluster_domain"
	"github.com/open-cluster-management/discovery/pkg/ocm/domain/subscription_domain"
	"github.com/open-cluster-management/discovery/pkg/ocm/services/auth_service"
	"github.com/open-cluster-management/discovery/pkg/ocm/services/cluster_service"
	"github.com/open-cluster-management/discovery/pkg/ocm/services/subscription_service"
)

var (
	getTokenFunc func(auth_domain.AuthRequest) (string, error)

	getClustersFunc func() ([]cluster_domain.Cluster, error)
	clusterGetter   = clusterGetterMock{}

	getSubscriptionsFunc func() ([]subscription_domain.Subscription, error)
	subscriptionGetter   = subscriptionGetterMock{}
)

// This mocks the authService request and returns a dummy access token
type authServiceMock struct{}

func (m *authServiceMock) GetToken(request auth_domain.AuthRequest) (string, error) {
	return getTokenFunc(request)
}

// The mocks the GetClusters request to return a select few clusters without connection
// to an external datasource
type clusterGetterMock struct{}

func (m *clusterGetterMock) GetClusters() ([]cluster_domain.Cluster, error) {
	return getClustersFunc()
}

// This mocks the NewClient function and returns an instance of the clusterGetterMock
type clusterClientGeneratorMock struct{}

func (m *clusterClientGeneratorMock) NewClient(config cluster_domain.ClusterRequest) cluster_service.ClusterGetter {
	return &clusterGetter
}

// The mocks the GetClusters request to return a select few clusters without connection
// to an external datasource
type subscriptionGetterMock struct{}

func (m *subscriptionGetterMock) GetSubscriptions() ([]subscription_domain.Subscription, error) {
	return getSubscriptionsFunc()
}

// This mocks the NewClient function and returns an instance of the subscriptionGetterMock
type subscriptionClientGeneratorMock struct{}

func (m *subscriptionClientGeneratorMock) NewClient(config subscription_domain.SubscriptionRequest) subscription_service.SubscriptionGetter {
	return &subscriptionGetter
}

// clustersResponse takes in a file with cluster data and returns a new mock function
func clustersResponse(testdata string) func() ([]cluster_domain.Cluster, error) {
	return func() ([]cluster_domain.Cluster, error) {
		file, _ := ioutil.ReadFile(testdata)
		clusters := []cluster_domain.Cluster{}
		err := json.Unmarshal([]byte(file), &clusters)
		return clusters, err
	}
}

// clustersResponse takes in a file with subscription data and returns a new mock function
func subscriptionResponse(testdata string) func() ([]subscription_domain.Subscription, error) {
	return func() ([]subscription_domain.Subscription, error) {
		file, _ := ioutil.ReadFile(testdata)
		subscriptions := []subscription_domain.Subscription{}
		err := json.Unmarshal([]byte(file), &subscriptions)
		return subscriptions, err
	}
}

func TestDiscoverClusters(t *testing.T) {
	type args struct {
		token   string
		baseURL string
		filters discoveryv1.Filter
	}
	tests := []struct {
		name             string
		authfunc         func(auth_domain.AuthRequest) (string, error)
		clusterFunc      func() ([]cluster_domain.Cluster, error)
		subscriptionFunc func() ([]subscription_domain.Subscription, error)
		args             args
		want             int
		wantErr          bool
	}{
		{
			name: "Three clusters and subscriptions",
			authfunc: func(auth_domain.AuthRequest) (string, error) {
				// this mock return a dummy token
				return "valid_access_token", nil
			},
			// this mock returns 3 clusters read from mock_clusters.json
			clusterFunc: clustersResponse("testdata/3_mock_clusters.json"),
			// this mock returns 3 subscriptions read from mock_subscriptions.json
			subscriptionFunc: subscriptionResponse("testdata/3_mock_subscriptions.json"),
			args: args{
				token:   "test",
				baseURL: "test",
				filters: discoveryv1.Filter{},
			},
			want:    3,
			wantErr: false,
		},
		{
			name: "Three clusters one subscription",
			authfunc: func(auth_domain.AuthRequest) (string, error) {
				// this mock return a dummy token
				return "valid_access_token", nil
			},
			// this mock returns 3 clusters read from mock_clusters.json
			clusterFunc: clustersResponse("testdata/3_mock_clusters.json"),
			// this mock returns 3 subscriptions read from mock_subscriptions.json
			subscriptionFunc: subscriptionResponse("testdata/1_mock_subscription.json"),
			args: args{
				token:   "test",
				baseURL: "test",
				filters: discoveryv1.Filter{},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "One clusters three subscription",
			authfunc: func(auth_domain.AuthRequest) (string, error) {
				// this mock return a dummy token
				return "valid_access_token", nil
			},
			// this mock returns 3 clusters read from mock_clusters.json
			clusterFunc: clustersResponse("testdata/1_mock_cluster.json"),
			// this mock returns 3 subscriptions read from mock_subscriptions.json
			subscriptionFunc: subscriptionResponse("testdata/3_mock_subscriptions.json"),
			args: args{
				token:   "test",
				baseURL: "test",
				filters: discoveryv1.Filter{},
			},
			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth_service.AuthClient = &authServiceMock{}                                          // Mocks out the call to auth service
			cluster_service.ClusterClientGenerator = &clusterClientGeneratorMock{}                // Mocks out the cluster client creation
			subscription_service.SubscriptionClientGenerator = &subscriptionClientGeneratorMock{} // Mocks out the subscription client creation

			getTokenFunc = tt.authfunc
			getClustersFunc = tt.clusterFunc
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
