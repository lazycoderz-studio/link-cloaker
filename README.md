# link-cloaker


## Testing Guide

### Running Application
```shell
go mod tidy
go run main.go
```

### Running Tests

#### Adding Links
 ```curl --location 'http://localhost:8080/update/abc' \
--header 'Content-Type: application/json' \
--data '{
    "real": "https://google.com",
    "bot": "https://facebook.com"
}'
```

change  your id, real and bot data

### Testing redirection
Open browser and navigate to http://localhost:8080/abc
change abc to your route

