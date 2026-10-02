

<!--
 Copyright (c) 2021 ActiveCHooN

 Este software se distribuye bajo la Licencia MIT.
 https://opensource.org/licenses/MIT
-->

# Gin GORM filter
![GitHub](https://img.shields.io/github/license/ActiveChooN/gin-gorm-filter)
![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/ActiveChooN/gin-gorm-filter/ci.yml?branch=master)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/ActiveChooN/gin-gorm-filter)

La función de ámbito (scope) para consultas de GORM proporciona un filtrado sencillo mediante parámetros de consulta

## Uso

```(shell)
go get github.com/ActiveChooN/gin-gorm-filter
```

## Definición del modelo
```go
type UserModel struct {
    gorm.Model
    Username string `gorm:"uniqueIndex" filter:"param:login;searchable;filterable"`
    FullName string `filter:"searchable"`
    Role     string `filter:"filterable"`
}
```
`param` en este caso define un nombre de columna personalizado para el parámetro de consulta

## Ejemplo de controlador
```go
func GetUsers(c *gin.Context) {
	var users []UserModel
	var usersCount int64
	var params filter.QueryParams

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	err := db.Model(&UserModel{}).Scopes(
		filter.FilterByQueryParams(params, filter.ALL),
	).Count(&usersCount).Find(&users).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	serializer := serializers.PaginatedUsers{Users: users, Count: usersCount}
	c.JSON(http.StatusOK, serializer.Response())
}
```
Aquí se puede utilizar cualquier combinación de filtros, como `filter.PAGINATE|filter.ORDER_BY`, por ejemplo. **Nota importante:** El modelo de GORM debe inicializarse primero para la base de datos; de lo contrario, el filtrado y la búsqueda no funcionarán

## Ejemplo de solicitud
```(shell)
curl -X GET http://localhost:8080/users?page=1&page_size=10&order_by=username&order_direction=asc&filter="name:John"
```

## Operadores de filtro admitidos
- :   El operador de igualdad `filter=username:John` coincide únicamente cuando el nombre de usuario es exactamente `John`
- \>  El operador mayor que `filter=age>35` coincide únicamente cuando la edad es mayor a 35
- \<  El operador menor que `filter=salary<80000` coincide únicamente cuando el salario es menor a 80.000
- \>= El operador mayor o igual que `filter=items>=100` coincide únicamente cuando los elementos son al menos 100
- \<= El operador menor o igual que `filter=score<=100000` coincide cuando la puntuación es 100.000 o inferior
- \!= El operador distinto de `state!=FAIL` coincide cuando el estado tiene cualquier valor distinto de FAIL
- \~  El operador de búsqueda (like) `filter=lastName~illi` coincide cuando lastName contiene la subcadena `illi`

## Lista de tareas
- [x] Escribir pruebas para la librería con integración CI
- [x] Agregar soporte para búsqueda insensible a mayúsculas y minúsculas
- [x] Agregar otros filtros, como > o !=
- [x] Agregar soporte para múltiples filtros en una sola consulta
- [ ] Agregar soporte para filtrar modelos relacionados
