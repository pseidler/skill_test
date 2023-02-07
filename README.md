# API

The aim of our solution is to create an API service as a gateway to ingest a data with high volume, variety at high velocity into a broker.

For this implemention test, we start very simple. We create a basic version of an API service that enables a customer or integrator to quickly send data to a custom endpoint. The service handles the call and forwards it to a broker, for example to Azure Eventhub. The destination broker is specified as a field in the message. 

There are three main requirements for this service and one optional.

1. Design an API for a dummy event of your choice. You can chose the endpoint type and protocols (REST, GRPC, HTTP1/2/3 etc) as well as the message format and envelope and/or payload and their fields.
2. Implement the async endpoint showing some understanding of DRY and SOLID principles. You do not have to implement the (customer) client connection and neither the message handling to the Azure Eventhub, a mock is sufficient; but of cource go as far as you want with your implementation.
3. Show a proper, coherent testing approach; you can use mocks, it can but does not have to be full TDD
4. Optional: Containerisation
5. Optional: Handling of large data points, for example image data

Minimal non-functional requirements: 
 - Handle high volume of data
 - Handle high velocity of data

Some more help:

1. Use a professional approach of planning, design, implementation, document it (README please), and show in running tests/benchmarks that your solution does what it should. There is no right or wrong here, but be prepared to comment on and argue for the choices that you have made.
2. Treat this like an agile project and set yourself a timeframe to deliver this task and stick to it. The most important aspect is to deliver something in you version 1 that works and can be extended


# FAQ

1. **Should the dummy event be an empty object passed to the API, or should I listen to a queue to some event which will contain data?**

    In reality, it is most likely that there will be a customer client queing data and sending them to the API (push). For this test, just create your own event for test purposes; no subscription to any queue needed.

2. **High volume of data - in this case, does it mean that this dummy object will be large itself, or it will be a small portion of data called with high velocity?**

    Payloads will contain mostly IoT events so are relatively small. Even for other larger data sets from DBs, the data points themselves will not be too big, so the emphasis is on velocity. There is the option to implement the API for handling image data, so that would be large data points at high velocity.


3. **Do we care about consistency of data, i.e. should data be delivered all in one or with the possibility to track dependencies?**

    You can implement bulk or batch endpoints if you see a need as by your design. Order of events will not have to be kept. A batch can contain different operations, but it is not necessary to show this during this test.

4. **Do we have any maximumWaitTime for a response from Azure Event Hub specified?**

    For these kind of questions, make your own decision and log the decision for your design/implementation.

5. **Do we have the possibility to configure Azure EventHub settings (e.g. partition keys) or shall we only mock it for purpose of this task?**

    For this task, the choice is with you, you can mock it fully or even provide a full integration with Azure natively/through the SDK. Whatever you do, log your decision if you think that it has an impact on the requirements or working of the API.