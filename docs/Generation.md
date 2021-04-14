## Test Generation Details

There are now two generation scripts which can now be used to create a large number of randomly named mock discoveredclusters and a subset of managed clusters, all of which are used in our mock server. These scripts are located at:
`testserver/generate.go`

## Generate Discovered Clusters

In order to generate the json which can be used to generate large numbers of mock discovered clusters, a sample command would be:
`./testserver/generate.go -tot=999 -d=24`
In this case the "tot" will represent the total number of clusters we are generating. The "d" indicates a time period for randomly generating the used by each cluster. If we pass 24 for this value, every generated date will be within the last 24 days from the time the script is run.
The output of the script is a file located at: `testserver/data/scenarios/onek_clusters/subscription_response.json`
It takes the form of a json outline for a SubscriptionList which is recognized by the mock server and the discovery operator.

In order to point the mock server to this new list of discovered clusters, we need to edit the mock server deployment so that the scenario is equal to "onekClusters"

## Generate Managed Clusters

In order to generate a large number of mock managed clusters, a sample command would be:
`./testserver/generate_managed.go -tot=500`
Again, the "tot" will indicate the total number of clusters we are generating. There is a validation to ensure that the number of managed clusters is fewer than the number of discovered clusters in our SubscriptionList json. These managed clusters take the form of buffer separated yaml which use the cluster IDs found in our SubscriptionList json. Because these id's will match, once we have created the managed clusters, a label will be applied to the corresponding discovered cluster indicating that it is managed. In order for the managed clusters to actually be created, we will use:
`oc apply -f testserver/data/sample_managed_clusters.yaml`


## Testing

Once we have set the scenario to our new list of discoveredclusters and managed clusters, we can use the following command to check that the expected number of clusters are being managed:
`oc get discoveredcluster -o yaml | grep 'isManagedCluster: "true"' | wc -l`