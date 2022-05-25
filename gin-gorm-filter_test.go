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

type User struct {
	Id       int64
	Username string `filter:"param:login;searchable;filterable"`
	FullName string `filter:"param:name;searchable"`
	Email    string `filter:"filterable"`
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

// TestFiltersBasic is a test suite for basic filters functionality.
func (s *TestSuite) TestFiltersBasic() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "filter=login:sampleUser",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "Username" = \$1`).
		WithArgs("sampleUser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "Username", "FullName", "Email", "Password"}))
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
	s.mock.ExpectQuery(`^SELECT \* FROM "users" ORDER`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "Username", "FullName", "Email", "Password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, ALL)).Find(&users).Error
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
		WillReturnRows(sqlmock.NewRows([]string{"id", "Username", "FullName", "Email", "Password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, SEARCH)).Find(&users).Error
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

	s.mock.ExpectQuery(`^SELECT \* FROM "users" WHERE \("Username" LIKE \$1 OR "FullName" LIKE \$2\)`).
		WithArgs("%John%", "%John%").
		WillReturnRows(sqlmock.NewRows([]string{"id", "Username", "FullName", "Email", "Password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, ALL)).Find(&users).Error
	s.NoError(err)
}

// TestFiltersPaginateOnly is a test suite for pagination functionality.
func (s *TestSuite) TestFiltersPaginateOnly() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "page=2&per_page=10",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" ORDER BY "id" DESC LIMIT 10 OFFSET 10$`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "Username", "FullName", "Email", "Password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, ALL)).Find(&users).Error
	s.NoError(err)
}

// TestFiltersOrderBy is a test suite for order by functionality.
func (s *TestSuite) TestFiltersOrderBy() {
	var users []User
	ctx := gin.Context{}
	ctx.Request = &http.Request{
		URL: &url.URL{
			RawQuery: "order_by=Email&desc=false",
		},
	}

	s.mock.ExpectQuery(`^SELECT \* FROM "users" ORDER BY "Email"$`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "Username", "FullName", "Email", "Password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, ORDER_BY)).Find(&users).Error
	s.NoError(err)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
