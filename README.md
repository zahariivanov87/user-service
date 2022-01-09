# user-service

Simple user service is meant to demonstrate [Clean Architecture ](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) code style of Uncle Bob!  The idea is to separate core abstraction from low level details. 


## capabilities

- User creation
- User modification
- All users retrieval (paginated)
- User deletion

## Layers
 Service is divided on the following layers:
 
 - Outer layer - Controllers (userview package) handles HTTP requests.
 - Business logic is represented by (user package) internal/user/manager. All business rules (Use Cases) are defined inside.
 - Domain models are wrapped in core package.

According to Clean Architecture guidances outer layers can go into inner layers only through interfaces. For example:
CreateUserEndpoint does not depend directly to user manager but interface:
```
type UserCreator interface {
	CreateUser(ctx context.Context, user core.User) error
}
```

## tests

Tests are using Ginkgo/Gomega -> https://onsi.github.io/ginkgo/
```
cd internal/user
ginkgo
```


## how to run

From root dir execute:
```
docker-compose up --build
```

## Example API requests

Create user:
```
curl --request POST \
  --url http://localhost:8080/api/public/v1/users \
  --header 'Content-Type: application/json' \
  --data '{
"username": "zahari",
"first_name": "Zahari",
"last_name": "Ivanov",
"password": "password",
"nickname": "Harry",
"email": "harry@gmail.com",
"country": "BG"
}'
```

Get All users:
```
curl --request GET \
  --url 'http://localhost:8080/api/public/v1/users?limit=50' \
  --header 'Content-Type: application/json'

```

## room for improvement / next steps

- test for all layers - view & db layer (output port)
- healthcheck
- k8s config
