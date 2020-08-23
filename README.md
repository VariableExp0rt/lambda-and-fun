# Lambda Over HTTPS to Create EKS Clusters

This idea was initially just to identify how I might start using the new Serverless Application Model, however, I was curious how far I could go using the aws-sdk-go library to create everything I needed to start a cluster.

`go build -o main main.go`

`export ACCOUNT="Account ID"`

`./main --file-path="deployment.zip"`

You can check all the resources are created by signing into the console (though I've tried to have messages print upon some successful creation of a resource), I think I will expand this to have other Lambda functions be triggered by the successful deployment of a cluster, to then do some kind of init script configuration magic or start some necessary workloads.

I added a small test at the end of the main function of this program, which really should be separated out into it's own unit testing file but I'll handle that at a later date when I have read more about proper testing lifecycles.

##### Refactoring

I should probably have made a struct with methods of the services I'm using such as session, the services, so then I can user a pointer to the struct to be able to consume these within functions (?). This is the next step in making it a bit more readable.

Example for my reference;

```
type Libraries struct {
    session *session.Session
    lmbsvc *lambda.New
}
```

Improve error handling and separate each creation of a resource into it's own function to give more flexbility on the CLI to skip role or other resource creation.

Reference when going through setting up a Cluster via Lambda;

[ ] eksctl.io

[ ] [AWS EKS User Guide](https://docs.aws.amazon.com/eks/latest/userguide/getting-started.html)

The idea is to have the cluster configuration in a separate package, and import into the Lambda to be used with 'CreateCluster'

##### TODO

[ ] Write tests
[ ] Improve error handling
[ ] Refactor huge configuration function
[ ] Create TestMethod subcommand
