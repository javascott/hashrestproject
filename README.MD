**Welcome to the HashedPassword Web Server**

To Run:
- go run server.go

Testing POST:
- I ran testing using Postman:
- send POST to localhost:8080/hash
- body select "Raw" radio box and enter password=XXXX
- it will return the unique identifier {id}

Testing GET:
- send GET to localhost:8080/hash/{id}
- it will return the hashed password after the specified time


