# user-service

Simple user service is meant to demonstrate [Clean Architecture ](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) code style of Uncle Bob!

## capabilities

- User creation
- User modification
- All users retrieval (paginated)
- User deletion

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
