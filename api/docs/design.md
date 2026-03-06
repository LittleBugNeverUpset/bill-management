# 账单管理系统设计

### 用户
字段	类型	说明
id	bigint	主键，自增
username	varchar(50)	用户名，唯一
password	varchar(100)	密码（bcrypt 加密）
nickname	varchar(50)	昵称
email	varchar(100)	邮箱，可唯一
role	varchar(20)	角色：admin /user
status	tinyint	1 正常 0 禁用
created_at	datetime	创建时间
updated_at	datetime	更新时间
deleted_at	datetime	软删除


```
写 model/user.go（结构体 + 表 + 钩子）
写 dao/user_dao.go（数据库 CURD）
写 service/user_service.go（注册、登录、登出）
写 controller/user_controller.go（HTTP 接口）
路由注册 + 中间件集成
测试：注册 → 登录 → 鉴权 → 登出
```

### 账单

### 家庭模块（可选）


