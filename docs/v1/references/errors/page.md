## Errors

Gofr has set of predefined errors which also sets status code as per REST API guidelines.

### DB

It is used to display an error which arises when communicating with a database (such as incorrect query syntax, incorrect values passed for arguments). e.g, if a request is sent to redis with a key which is not present in the database, the following error is displayed:

`redis: nil`

```go
// sets status code to 500 and error message as error provided
errors.DB{Err: sql.ErrNoRows}
```

### EntityAlreadyExists

It is used to display an error which arises while creating an existing entity.

```go
// sets the response status code as 200 and sends the object resp as the response body.
errors.EntityAlreadyExists{}
```

### EntityNotFound

It is used to display an error when an entity could not be found. e.g, when a GET request is sent to an invalid endpoint, the following error is displayed:

`No 'Route' found for Id: 'GET /'`

```go
// sets status code to 404 and error message : No '<entity>' found for Id: '<id>'
errors.EntityNotFound{Entity: "name", ID: "1"}
```

### InvalidParam

It is used to display an error when invalid parameters are passed in a request (such as incorrect latitude/longitude data). e.g, if a POST request is sent to postgres for creating a new record with errors in the body, such as incorrect datatype of a parameter, the following error is displayed:

`Incorrect value for parameter: body`

```go
// sets status code as 400 and error message to :Incorrect value for parameter: <param>
errors.InvalidParam{Param: []string{"name"}}
```

### MethodMissing

It is used to display an error when a method is utilised which has not been defined for an entity. (such as a **POST** request which has no associated handler).

```go
// sets status code to 405 and error message as Method 'POST' for 'dummy' not defined yet
errors.MethodMissing{Method: "POST",URL: "/dummy"}
```

### MissingParam

It is used to display an error when parameters are missed in a request (such as only the name is passed in a request where both name and ID will have to be passed). e.g, if we send a request to a database which requires userID and the name, and only provide the name in the body of the request, the following error is displayed:

`Parameter userID is required for this request`

```go
// sets status code as 400 and error message to :Missing value for parameter: <param>
errors.MissingParam{Param: []string{"name"}}
```

### MissingTemplate

It is used to display an error when a template could not be loaded for reasons such as incorrect file path, and the absence of the filename. e.g, if a template `defaultTemplate.html` is not present at the location `incorrect/location`, the following error is displayed:

`Template defaultTemplate.html does not exist at location: incorrect/location`

```go
// sets status code to 500 and error message : "Template <filename> does not exist at location: <FileLocation>"
errors.MissingTemplate{
	FileLocation: "example/test",
	FileName:     "example/test/test.go",
	}
```

### Raw

It is used to overide gofr's default error format and status code.

```go
// sets to given status code and populates different fields provided
// overrides gofr error response format
errors.Raw{
	StatusCode: 500,
	Err:        errors.New("database error"),
}
```

### MultipleErrors

It is used to display an error response which indicates that more than one error has been encountered in the request. ( e.g, passing a validation request for phone number, address, and email and more than one of the fields have incorrect values). e.g, if a request is sent with an incorrect `ID`, and is also missing a parameter `name`, the following error is displayed:

`Parameter name is required for this request
 Incorrect value for parameter: ID`

### Custom Error

The response is generated by populating the `Response` struct and displaying it:

```go
// sets to given status code and populates different fields provided
errors.Response{
	StatusCode: 500, // it represents the status code generated for the HTTP request.
	Code:       "ERR_INTERNAL_SERVER_ERROR", //  type of error based on the encountered status code.
	Reason:     "Internal Server Error", // cause of the error
	ResourceID: "1", // display the resource ID
	Detail:     errors.New("database error"), // Contains details about the cause of the error
	Path:       "example/sample-api", // refer to the location in the response
	// If defined, it must contain the Code and Reason fields defined above, along with any additional fields to provide more information if necessary.
	RootCauses: nil,
	DateTime:   errors.DateTime{}, // time at which the response was displayed
	}
```

| Element    | Datatype                 | Description                                                                                                                                            |
| ---------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| StatusCode | int                      | The element represents status code generated for the HTTP request.                                                                                     |
| Code       | string                   | It clarifies the type of error based on the status code encountered.                                                                                   |
| Reason     | string                   | It is used to display the cause of the error.                                                                                                          |
| ResourceID | string                   | It is used to display the resource ID.                                                                                                                 |
| Detail     | interface                | Contains details about the cause of the error                                                                                                          |
| Path       | string                   | It is used to refer the location in the response                                                                                                       |
| RootCauses | []map[string]interface{} | If defined, it must contain the `Code` and the `Reason` fields defined above along with any additional fields to provide more information if necessary |
| DateTime   | struct                   | It consists of Value(type **string**) and TimeZone(type **string**). It displays the time at which the response was displayed.                         |