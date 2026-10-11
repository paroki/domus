package test

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/unit/model"
	"github.com/paroki/domus/api/internal/platform/authz"
	"github.com/paroki/domus/api/testutil"
	"github.com/stretchr/testify/suite"
)

type ParishSuite struct {
	testutil.ApiTestSuite[model.ParishResponse]
}

func (s *ParishSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.GrantRole(authz.RoleSuperadmin, authz.System)
}

func (s *ParishSuite) createDiocese(name string) model.DioceseResponse {
	dReq := model.CreateDioceseRequest{Name: name}
	s.Request("/dioceses", fiber.MethodPost, dReq)
	s.Created()
	dResp := testutil.DecodeOther[model.DioceseResponse](s)
	return dResp.Data
}

func (s *ParishSuite) TestCreate_Success() {
	diocese := s.createDiocese("Keuskupan Agung Jakarta")

	req := model.CreateParishRequest{
		Name:      "Paroki Katedral Jakarta",
		DioceseID: &diocese.ID,
	}

	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Greater(res.Data.ID, 0)
	s.Equal("Paroki Katedral Jakarta", res.Data.Name)
	s.NotNil(res.Data.Diocese)
	s.Equal(diocese.ID, res.Data.Diocese.ID)
	s.Equal("Keuskupan Agung Jakarta", res.Data.Diocese.Name)
}

func (s *ParishSuite) TestCreate_WithoutDiocese_Success() {
	req := model.CreateParishRequest{
		Name: "Paroki Tanpa Keuskupan",
	}

	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()

	res := s.GetResponse()
	s.Greater(res.Data.ID, 0)
	s.Equal("Paroki Tanpa Keuskupan", res.Data.Name)
}

func (s *ParishSuite) TestCreate_ValidationError() {
	req := model.CreateParishRequest{
		Name: "",
	}

	s.Request("/parishes", fiber.MethodPost, req)
	s.AssertStatus(fiber.StatusUnprocessableEntity)
}

func (s *ParishSuite) TestGetAll_Success() {
	req1 := model.CreateParishRequest{Name: "Paroki Theresia"}
	s.Request("/parishes", fiber.MethodPost, req1)
	s.Created()

	req2 := model.CreateParishRequest{Name: "Paroki Blok B"}
	s.Request("/parishes", fiber.MethodPost, req2)
	s.Created()

	s.Request("/parishes", fiber.MethodGet, nil)
	s.OK()

	res := s.PagedResponse()
	s.GreaterOrEqual(len(res.Data), 2)
}

func (s *ParishSuite) TestGetByID_Success() {
	diocese := s.createDiocese("Keuskupan Bandung")

	req := model.CreateParishRequest{
		Name:      "Paroki Santo Petrus Bandung",
		DioceseID: &diocese.ID,
	}
	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/parishes/%d", created.ID), fiber.MethodGet, nil)
	s.OK()

	res := s.GetResponse()
	s.Equal(created.ID, res.Data.ID)
	s.Equal("Paroki Santo Petrus Bandung", res.Data.Name)
	s.NotNil(res.Data.Diocese)
	s.Equal(diocese.ID, res.Data.Diocese.ID)
}

func (s *ParishSuite) TestGetByID_NotFound() {
	s.Request("/parishes/999999", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *ParishSuite) TestGetByID_InvalidID() {
	s.Request("/parishes/invalid-id", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusBadRequest)
}

func (s *ParishSuite) TestUpdate_Success() {
	diocese1 := s.createDiocese("Keuskupan Surabaya")
	diocese2 := s.createDiocese("Keuskupan Malang")

	req := model.CreateParishRequest{
		Name:      "Paroki Hati Kudus Surabaya",
		DioceseID: &diocese1.ID,
	}
	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	updateReq := model.UpdateParishRequest{
		Name:      "Paroki Hati Kudus Yesus",
		DioceseID: &diocese2.ID,
	}
	s.Request(fmt.Sprintf("/parishes/%d", created.ID), fiber.MethodPut, updateReq)
	s.OK()

	res := s.GetResponse()
	s.Equal(created.ID, res.Data.ID)
	s.Equal("Paroki Hati Kudus Yesus", res.Data.Name)
	s.NotNil(res.Data.Diocese)
	s.Equal(diocese2.ID, res.Data.Diocese.ID)
}

func (s *ParishSuite) TestUpdate_NotFound() {
	updateReq := model.UpdateParishRequest{Name: "Non Existent Parish"}
	s.Request("/parishes/999999", fiber.MethodPut, updateReq)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *ParishSuite) TestDelete_Success() {
	req := model.CreateParishRequest{Name: "Paroki Santo Yoseph Denpasar"}
	s.Request("/parishes", fiber.MethodPost, req)
	s.Created()
	created := s.GetResponse().Data

	s.Request(fmt.Sprintf("/parishes/%d", created.ID), fiber.MethodDelete, nil)
	s.NoContent()

	s.Request(fmt.Sprintf("/parishes/%d", created.ID), fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *ParishSuite) TestDelete_NotFound() {
	s.Request("/parishes/999999", fiber.MethodDelete, nil)
	s.AssertStatus(fiber.StatusNotFound)
}

func (s *ParishSuite) TestForbidden_NoPermission() {
	// new user without any membership
	s.User.ID = core.GenerateID()

	s.Request("/parishes", fiber.MethodGet, nil)
	s.AssertStatus(fiber.StatusForbidden)
}

func TestParishSuite(t *testing.T) {
	suite.Run(t, new(ParishSuite))
}

