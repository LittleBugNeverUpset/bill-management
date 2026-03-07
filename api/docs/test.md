
``` bash
go clean -cache && go clean -modcache
go mod tidy
go run cmd/server/main.go


curl -X POST http://localhost:8080/api/user/login -H "Content-Type: application/json" -d '{
    "username": "test_user_02",
    "password": "123456"
}'

```


模拟创建订单
```bash
curl -X POST http://localhost:8080/bill -H "Content-Type: application/json" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxOSwidXNlcm5hbWUiOiJ0ZXN0X3VzZXJfMDIiLCJyb2xlIjoidXNlciIsImV4cCI6MTc3Mjg5NDUyOCwibmJmIjoxNzcyODkwOTI4LCJpYXQiOjE3NzI4OTA5Mjh9.kqgF7mU23k8XiY6diozgnMEi-wNTvt9Nern25YKsTOE" -d '{
    "category_id": 1,
    "amount": 100.50,
    "type": true,
    "remark": "测试账单"
}'

```


``` bash
curl -X GET "http://localhost:8080/bill?page=1&page_size=10" \
-H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxOSwidXNlcm5hbWUiOiJ0ZXN0X3VzZXJfMDIiLCJyb2xlIjoidXNlciIsImV4cCI6MTc3Mjg5NDUyOCwibmJmIjoxNzcyODkwOTI4LCJpYXQiOjE3NzI4OTA5Mjh9.kqgF7mU23k8XiY6diozgnMEi-wNTvt9Nern25YKsTOE"


```