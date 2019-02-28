README

jumpcloud is a server that meets the requirements of the JumpCloud Interview Assignment.



API:

POST /hash       - Calculates the base64 encoding of the sha512 hash of the provided input and will not complete the request within 5 seconds.  Accepts a JSON body of the form {"password": string}.  Returns a JSON body of the form {"hash": string}.  If "hash" is not specified, it defaults to the empty string.  Returns HTTP response code 200 upon success.

GET  /statistics - Retrieves statistics of the number of requests to "POST /hash" that the server has handled and the average time to process the requests in microseconds.  Returns a JSON body of the form {"total": int, "average": int}.  Returns HTTP response code 200 upon success.

POST /shutdown   - Initiates a graceful shutdown of the server, allowing in-flight requests to complete.  Returns HTTP response code 202 upon success.


Usage:

To install, issue "go get github.com/dkrisa/jumpcloud".  To run, issue "jumpcloud".  Go version 1.11 required.  The app listens on port 8000.


Design choices & idiosyncrasies:

To avoid excessive locking, calls to "/statistics" do not lock the underlying object when extracting the statistics.  As a result, it is possible that for a given call to "POST /hash", the "average" field has been updated but the "total" field has not.  The only potentially confusing scenario is for the very first transaction, when the "average" is set to a value, but the "total" is still at 0.

Since the statistics are updated in a goroutine that is kicked off while handling the request, the following scenarios are all possible but deemed acceptable since the impact of these scenarios should be relatively small on the overall statistics:
1. A "POST /hash" has been completed but its data is not yet reflected in the statistics.
2. A "POST /hash" fails after the encoded has has been computed and its data is included in the statistics.

The latency statistic is measured from the time the server begins serving the request to immediately after it writes the response.  The 5 second response delay begins when the server begins serving the request.

The request and response bodies for "POST /hash" are in JSON format for ease of use and standardization.
