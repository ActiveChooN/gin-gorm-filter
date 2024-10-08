// Copyright (c) 2022 ActiveChooN
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package filter

import (
	"database/sql"
	"net/http"
	"net/url"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Organization struct {
	Id   uint   `filter:"param:id;filterable"`
	Name string `filter:"param:name;searchable"`
}

type User struct {
	Id             uint   `filter:"param:id;filterable"`
	Username       string `filter:"param:login;searchable;filterable"`
	FullName       string `filter:"param:name;searchable"`
	Email          string `filter:"filterable"`
	OrganizationId uint
	Organization   Organization
	// This field is not filtered.
	Password string
}

type TestSuite struct {
	suite.Suite
	db   *gorm.DB
	mock sqlmock.Sqlmock
}

func (s *TestSuite) SetupTest() {
	var (
		db  *sql.DB
		err error
	)

	db, s.mock, err = sqlmock.New()
	s.NoError(err)
	s.NotNil(db)
	s.NotNil(s.mock)

	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	s.db, err = gorm.Open(dialector, &gorm.Config{})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), s.db)
}

func (s *TestSuite) TearDownTest() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	db.Close()
}

// TestFiltersBasic is a test for basic filters functionality.
func (s *TestSuite) TestFiltersBasic() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=login:sampleUser",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"."username" = \$1 ORDER BY "id" DESC LIMIT \$2$`).
		WithArgs("sampleUser", 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, ALL)).Find(&users).Error
	s.NoError(err)
}

// TestFiltersBasic is a test for basic filters functionality.
func (s *TestSuite) TestFiltersLike() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=login~samp",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"."username" LIKE \$1 ORDER BY "id" DESC LIMIT \$2$`).
		WithArgs("samp", 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, ALL)).Find(&users).Error
	s.NoError(err)
}

// Filtering for a field that is not filtered should not be performed
func (s *TestSuite) TestFiltersNotFilterable() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=password:samplePassword",
		},
	}
	s.mock.ExpectQuery(`^SELECT \* FROM "users"$`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, FILTER)).Find(&users).Error
	s.NoError(err)
}

// Filtering would not be applied if no config is provided.
func (s *TestSuite) TestFiltersNoFilterConfig() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=login:sampleUser",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users"$`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, SEARCH)).Find(&users).Error
	s.NoError(err)
}

// Filtering would not be applied if no config is provided.
func (s *TestSuite) TestFiltersNotEqualTo() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=id!=22",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"."id" <> \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, FILTER)).Find(&users).Error
	s.NoError(err)
}

func (s *TestSuite) TestFiltersLessThan() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=login<Phil",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"."username" < \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, FILTER)).Find(&users).Error
	s.NoError(err)
}

// Filtering would not be applied if no config is provided.
func (s *TestSuite) TestFiltersLessThanOrEqualTo() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=id<=200",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"."id" <= \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, FILTER)).Find(&users).Error
	s.NoError(err)
}

func (s *TestSuite) TestFiltersGreaterThan() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=id>100",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"."id" > \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, FILTER)).Find(&users).Error
	s.NoError(err)
}

func (s *TestSuite) TestFiltersGreaterThanOrEqualTo() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=id>=99",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"."id" >= \$1`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, FILTER)).Find(&users).Error
	s.NoError(err)
}

// TestFiltersSearchable is a test suite for searchable filters functionality.
func (s *TestSuite) TestFiltersSearchable() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "search=John",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" WHERE \(LOWER\("users"."username"\) LIKE \$1 OR LOWER\("users"."full_name"\) LIKE \$2\)$`).
		WithArgs("%john%", "%john%").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, SEARCH)).Find(&users).Error
	s.NoError(err)
}

// TestFiltersPaginateOnly is a test for pagination functionality.
func (s *TestSuite) TestFiltersPaginateOnly() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "page=2&per_page=10",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" ORDER BY "id" DESC LIMIT \$1 OFFSET \$2$`).
		WithArgs(10, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, ALL)).Find(&users).Error
	s.NoError(err)
}

// TestFiltersOrderBy is a test for order by functionality.
func (s *TestSuite) TestFiltersOrderBy() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "order_by=Email&order_direction=asc",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" ORDER BY "Email"$`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, ORDER_BY)).Find(&users).Error
	s.NoError(err)
}

// TestFiltersAndSearcg is test for filtering and searching simultaneously.
func (s *TestSuite) TestFiltersAndSearch() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=login:sampleUser&search=John",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" WHERE \(LOWER\("users"."username"\) LIKE \$1 OR LOWER\("users"."full_name"\) LIKE \$2\) AND "users"."username" = \$3$`).
		WithArgs("%john%", "%john%", "sampleUser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))

	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, FILTER|SEARCH)).Find(&users).Error
	s.NoError(err)
}

// TestFiltersMultipleColumns is a test for filtering on multiple columns.
func (s *TestSuite) TestFiltersMultipleColumns() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=login:sampleUser&filter=email:john@example.com",
		},
	}

	s.mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"."username" = \$1 AND "users"."email" = \$2$`).
		WithArgs("sampleUser", "john@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))

	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, FILTER)).Find(&users).Error
	s.NoError(err)
}

// TestFiltersWithJoin is a test for filtering with join.
func (s *TestSuite) TestFiltersWithJoin() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=id!=22",
		},
	}

	s.mock.ExpectQuery(`SELECT "users"."id","users"."username","users"."full_name","users"."email","users"."organization_id","users"."password","Organization"."id" AS "Organization__id","Organization"."name" AS "Organization__name" FROM "users" LEFT JOIN "organizations" "Organization" ON "users"."organization_id" = "Organization"."id" WHERE "users"."id" <> \$1$`).
		WithArgs("22").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "full_name", "email", "password"}))

	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, FILTER)).Joins("Organization").Find(&users).Error
	s.NoError(err)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
