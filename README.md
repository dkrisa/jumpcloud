README

jumpcloud is a server that meets the requirements of the JumpCloud Interview Assignment.



API:

POST /hash       - Calculates the base64 encoding of the sha512 hash of the provided input.  Accepts a JSON body of the form {"password": string}.  Returns a JSON body of the form {"hash": string}.  Returns HTTP response code 200 upon success.

GET  /statistics - Retrieves statistics of the number of requests the server has handled and the average time to process the requests in microseconds.  Returns a JSON body of the form {"total": int, "average": int}.  Returns HTTP response code 200 upon success.

POST /shutdown   - Initiates a graceful shutdown of the server, allowing in-flight requests to complete.  Returns HTTP response code 202 upon success.


Usage:

To run from source, clone this project into your  and issue "go run main.go" in the base directory.  Go version 1.11 required.
