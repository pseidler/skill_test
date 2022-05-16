# API

The aim of our solution is to create an API service as a gateway to ingest a lot of data fast into some broker.

We start very simple. We create a basic version of an API service that enables a customer or integrator to quickly send data to a custom endpoint. The service handles the call and forwards it to a broker, for example to Azure Eventhub. The destination broker is specified as a field in the message. 

There are three main requirements for this service and one optional.

1. Design an API for a dummy event. You can chose the endpoint type and protocols (REST, GRPC, HTTP1/2/3 etc)
2. Implement the endpoint (async if suitable) showing some understanding of DRY and SOLID principles. You do not have to implement the client connection and message handling to the Azure Eventhub, a mock is sufficient.
3. Show a proper, coherent testing approach; you can use mocks, it can but does not have to be full TDD
4. Optional: Containerisation

Some more help:

1. Use a professional approach of planning, design, implementation, document it (README please), and show in running tests/benchmarks that your solution does what it should. There is no right or wrong here, but be prepared to comment on and argue for the choices that you have made.
2. Treat this like an agile project and set yourself a timeframe to deliver this task and stick to it. The most important aspect is to deliver something in you version 1 that works and can be extended
