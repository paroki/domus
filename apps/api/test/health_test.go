package test

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/suite"
)

type HealthTestSuite struct {
	testutil.ApiTestSuite[core.AuthenticatedUser]
}

func (s *HealthTestSuite) TestPing_Success() {
	s.Request("/ping", fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	s.Equal(s.User.ID, res.Data.ID)
	s.Equal(s.User.Name, res.Data.Name)
	s.Equal(s.User.Email, res.Data.Email)
	s.Equal(s.User.WorkspaceID, res.Data.WorkspaceID)
	s.Equal(s.User.WorkspaceName, res.Data.WorkspaceName)
	s.Equal(s.User.WorkspaceRoles, res.Data.WorkspaceRoles)
}

func (s *HealthTestSuite) TestPing_Unauthorized() {
	s.RequestWithToken("/ping", fiber.MethodGet, nil, "")
	s.AssertStatus(fiber.StatusUnauthorized)
}

func TestHealthSuite(t *testing.T) {
	suite.Run(t, new(HealthTestSuite))
}
