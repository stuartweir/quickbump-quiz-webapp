*All this is subject to change*

# HTTP API

## GET

### /question/&lt;id&gt;

This will fetch the model data for the `Question` for the given `id`. Response will be JSON encoded. See the data model documentation for structure of the object sent in the response.

#### Example

    $ curl -v localhost:8888/api/question/1
    * About to connect() to localhost port 8888 (#0)
    *   Trying ::1...
    * Connected to localhost (::1) port 8888 (#0)
    > GET /api/question/1 HTTP/1.1
    > User-Agent: curl/7.29.0
    > Host: localhost:8888
    > Accept: */*
    >
    < HTTP/1.1 200 OK
    < Date: Sat, 30 Mar 2013 17:49:07 GMT
    < Transfer-Encoding: chunked
    < Content-Type: text/plain; charset=utf-8
    <
    * Connection #0 to host localhost left intact
    {"Created":"2013-03-30T10:47:29.137444-07:00","End":"2013-12-12T01:02:03Z","Data":{"Mode":"TEXT","Info":{"Question":"Why am I even doing this? I have papers to write and nobody is going to read this ...","WordLimit":42,"CharLimit":0}}}%                     

### /answer/&lt;id&gt;

This will fetch the model data for the `Answer` for the given `id`. Response will be JSON encoded. See the data model documentation for structure of the object sent in the response.

#### Example

    $ curl -v localhost:8888/api/answer/2
    * About to connect() to localhost port 8888 (#0)
    *   Trying ::1...
    * Connected to localhost (::1) port 8888 (#0)
    > GET /api/answer/2 HTTP/1.1
    > User-Agent: curl/7.29.0
    > Host: localhost:8888
    > Accept: */*
    >
    < HTTP/1.1 200 OK
    < Date: Sat, 30 Mar 2013 17:55:23 GMT
    < Transfer-Encoding: chunked
    < Content-Type: text/plain; charset=utf-8
    <
    * Connection #0 to host localhost left intact
    {"Question":"1","Created":"2013-03-30T10:52:49.45213-07:00","Response":"For great justice!"}

### If something goes wrong

If you make a request for a non-existent record, you should receive a response with a 404 status code and "Record not found" in the response body.

## POST

### /question

This will create a `Question` record. The request data should be JSON encoded. See the data model documentation for structure of the request. The response body will contain the id of the newly created record.

#### Example

    $ curl -vX POST localhost:8888/api/question -d '{
      "End": "2013-12-12T01:02:03.0Z",
      "Data": {
        "Mode": "TEXT",
        "Info": {
          "Question": "Why am I even doing this? I have papers to write and nobody is going to read this ...",
          "WordLimit": 42,
          "CharaterLimit": 0
         }
       }
    }'
    * About to connect() to localhost port 8888 (#0)
    *   Trying ::1...
    * Connected to localhost (::1) port 8888 (#0)
    > POST /api/question HTTP/1.1
    > User-Agent: curl/7.29.0
    > Host: localhost:8888
    > Accept: */*
    > Content-Length: 250
    > Content-Type: application/x-www-form-urlencoded
    >
    * upload completely sent off: 250 out of 250 bytes
    < HTTP/1.1 200 OK
    < Date: Sat, 30 Mar 2013 17:47:29 GMT
    < Transfer-Encoding: chunked
    < Content-Type: text/plain; charset=utf-8
    <
    * Connection #0 to host localhost left intact
    1%                                             

### /answer

This will create an `Answer` record. The request data should be JSON encoded. See the data model documentation for structure of the request. The response body will contain the id of the newly created record.

#### Example

    $ curl -vX POST localhost:8888/api/answer -d '{
      "QuestionId": "1",
      "Response": "For great justice!"
    }'
    * About to connect() to localhost port 8888 (#0)
    *   Trying ::1...
    * Connected to localhost (::1) port 8888 (#0)
    > POST /api/answer HTTP/1.1
    > User-Agent: curl/7.29.0
    > Host: localhost:8888
    > Accept: */*
    > Content-Length: 59
    > Content-Type: application/x-www-form-urlencoded
    >
    * upload completely sent off: 59 out of 59 bytes
    < HTTP/1.1 200 OK
    < Date: Sat, 30 Mar 2013 17:52:49 GMT
    < Transfer-Encoding: chunked
    < Content-Type: text/plain; charset=utf-8
    <
    * Connection #0 to host localhost left intact
    2%            

### If something goes wrong

If the record cannot be created - which should happen if your request data is malformed - a the response status code will probably be 500 and some explanation should be given in the response body.

## QUERY

Okay, QUERY is not an HTTP method. This section describes the 'QUERY' action but the method should be "GET'

Query filters should be encoded in the query string for the URL. Keys in the query string should match field names. Don't specify more than one value for a field. Filtering on fields that aren't top-level is unsupported.

A successful query will return a JSON object with the top-level keys being ids and their values being the model for the record with that id.

#### Example

    $ curl -v localhost:8888/api/answer\?Question=1
    * About to connect() to localhost port 8888 (#0)
    *   Trying ::1...
    * Connected to localhost (::1) port 8888 (#0)
    > GET /api/answer?Question=1 HTTP/1.1
    > User-Agent: curl/7.29.0
    > Host: localhost:8888
    > Accept: */*
    >
    < HTTP/1.1 200 OK
    < Date: Sat, 30 Mar 2013 18:25:21 GMT
    < Transfer-Encoding: chunked
    < Content-Type: text/plain; charset=utf-8
    <
    * Connection #0 to host localhost left intact
    {"2":{"Question":"1","Created":"2013-03-30T11:25:03.899316-07:00","Response":"For great justice!"}}% 


### If something goes wrong

If the query cannot be made, the response status code will probably be 500 and some explanation should be given in the response body.
