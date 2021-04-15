## Test Generation Details

There are now two generation scripts which can now be used to create a Subscriptions API response with a large number of randomly named clusters and a subset of managed cluster yamls objects, all of which are used in our mock server. These scripts are located at:
`testserver/generate-scripts`

## Generate Subscription Response 

In order to generate the json which can be used to generate a mock subscription response with a large numbers of clusters, a sample command would be:
`go run ./testserver/generate-scripts/generatesubscriptions/generate.go -tot=999 -d=24 -output=example-subscription-response.json`
In this case the "tot" will represent the total number of clusters we are generating. The "d" indicates a time period for randomly generating the used by each cluster. If we pass 24 for this value, every generated date will be within the last 24 days from the time the script is run.
The output file is written to a file specified via the "-output" flag.
It takes the form of a json outline for a SubscriptionList which is recognized by the mock server and the discovery operator.
## Generate Managed Clusters

In order to generate a large number of mock managed clusters, a sample command would be:
`go run ./testserver/generate-scripts/generatemanagedclusters/generate.go -tot=500 -input=example-subscription-response.json -output=example-managed-clusters.yaml`
Again, the "tot" will indicate the total number of clusters we are generating. There is a validation to ensure that the number of managed clusters is fewer than the number of discovered clusters in our SubscriptionList json. These managed clusters take the form of buffer separated yaml which use the cluster IDs found in our SubscriptionList json (passed in via the `-input` flag). Because these id's will match, once we have created the managed clusters, a label will be applied to the corresponding discovered cluster indicating that it is managed. In order for the managed clusters to actually be created, we will use:
`oc apply -f <output-file-from-generate-managed>`


## Testing

Once we have set the scenario to our new list of discoveredclusters and managed clusters, we can use the following command to check that the expected number of clusters are being managed:
`oc get discoveredcluster -o yaml | grep 'isManagedCluster: "true"' | wc -l`