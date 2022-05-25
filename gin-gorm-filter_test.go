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
	"github.com/stretchr/testify/assert"
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

func (s *TestSuite) SetupSuite() {
	var (
		db  *sql.DB
		err error
	)

	db, s.mock, err = sqlmock.New()
	require.NoError(s.T(), err)
	require.NotNil(s.T(), db)
	require.NotNil(s.T(), s.mock)

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

func (s *TestSuite) TearDownSuite() {
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
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Username", "FullName", "Email", "Password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, ALL)).Find(&users).Error
	assert.NoError(s.T(), err)
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
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Username", "FullName", "Email", "Password"}))
	err := s.db.Model(&User{}).Scopes(FilterByQuery(&ctx, ALL)).Find(&users).Error
	assert.NoError(s.T(), err)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
