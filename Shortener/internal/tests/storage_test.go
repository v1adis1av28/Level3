package tests

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/v1adis1av28/level3/shortener/internal/config"
	"github.com/v1adis1av28/level3/shortener/internal/storage"
	"github.com/v1adis1av28/level3/shortener/internal/testhelpers"
)

type StorageTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	storage     *storage.Storage
	ctx         context.Context
}

func (suite *StorageTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer
	dbConf := &config.DBConfig{MasterDbUrl: pgContainer.ConnectionString, SlaverUrl: []string{pgContainer.ConnectionString}}
	repository, err := storage.New(dbConf)
	if err != nil {
		log.Fatal(err)
	}
	suite.storage = repository
}

func (suite *StorageTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func (suite *StorageTestSuite) TestCreateCustomer() {
	t := suite.T()

	err := suite.storage.ShortenURL("https://abc.com", "ccc")
	assert.NoError(t, err)
	err = suite.storage.ShortenURL("https://def.com", "def")
	assert.NoError(t, err)
	err = suite.storage.ShortenURL("https://def.com", "ccc")
	assert.Error(t, err)
}

func (suite *StorageTestSuite) TestGetCustomerByEmail() {
	t := suite.T()

	url, err := suite.storage.GetURL("abc")
	assert.NoError(t, err)
	assert.NotNil(t, url)
	assert.Equal(t, "https://abc.com", url)

	url, err = suite.storage.GetURL("nonexistent")
	assert.Error(t, err)
	assert.Empty(t, url)
}

func (suite *StorageTestSuite) TestGetAnalytic() {
	t := suite.T()

	response, err := suite.storage.GetAnalytic("abc")
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.GreaterOrEqual(t, len(response.CurrentDayGroup), 1)
	assert.GreaterOrEqual(t, len(response.LastMonthGroup), 1)
	assert.GreaterOrEqual(t, len(response.UserAgentGroup), 1)
	assert.GreaterOrEqual(t, response.SummaryRedirectCount, 1)
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}
