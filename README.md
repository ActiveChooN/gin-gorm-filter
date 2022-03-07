<!--
 Copyright (c) 2021 ActiveCHooN

 This software is released under the MIT License.
 https://opensource.org/licenses/MIT
-->

# Gin GORM filter

Scope function for GORM queries provides easy filtering with query parameters

Fork dedicated to the usage of WiiSmile
## Usage

```(shell)
go get github.com/magellancl/gin-gorm-filter
```

## Model definition
```go
type UserModel struct {
    gorm.Model
    Username string `gorm:"uniqueIndex" filter:"param:login;searchable;filterable"`
    FullName string `filter:"searchable"`
    Role     string `filter:"filterable"`
}
```
`param` tag in that case defines custom column name for the query param

## Controller Example
```go
func GetUsers(c *gin.Context) {
	var users []UserModel
	var usersCount int64
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	err := db.Model(&UserModel{}).Scopes(
		filter.FilterByQuery(c, filter.ALL),
	).Count(&usersCount).Find(&users).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	serializer := serializers.PaginatedUsers{Users: users, Count: usersCount}
	c.JSON(http.StatusOK, serializer.Response())
}
```
Any filter combination can be used here `filter.PAGINATION|filter.ORDER_BY` e.g. **Important note:** GORM model should be initialize first for DB, otherwise filter and search won't work

## TODO list
- [ ] Write tests for the lib with CI integration
- [ ] Add ILIKE integration for PostgreSQL database
- [ ] Add other filters, like > or !=
