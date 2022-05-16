# API

We assume we have a deployed infrastructure around a Kubernetes cluster. The infrastructure takes care of security, in Azure for example by an Azure API Gateway and Istio with proper rules setup and OAuth. The trust boundary ends with Istio. 

The aim of our solution is to ingest a lot of data fast into different pipelines that are already deployed in the infrastructure. Pipelines can be a combination of a broker like the Azure Eventhub, other deployed services and lambda functions such as Azure functions with data ending up in databases, indices etc. 

We start very simple. The aim of this task is to create a basic version of an API service to be deployed into Kubernetes that enables a customer or integrator to quickly send data to any of the above mentioned pipelines through one or multiple endpoints. There are three main requirements for this service.

1. Design an API for a dummy Cloud Event (https://cloudevents.io/). You can chose the endpoint type and protocols (REST, GRPC, HTTP1/2/3 etc)
2. Implement the endpoint (async if suitable) showing some understanding of DRY and SOLID principles
3. Show a proper, coherent testing approach,;you can use mocks, it can but does not have to be full TDD
4. Optional: Containerisation

Some more help:

1. Use a professional approach of planning, design, implementation, document it (README please), and show in running tests/benchmarks that your solution does what it should. There is no right or wrong here, but be prepared to comment on and argue for the choices that you have made.
2. Treat this like an agile project and. Set yourself a timeframe to deliver this task and stick to it. The most important is to deliver something in you version 1 that works and can be extended
