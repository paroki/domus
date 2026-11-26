package service_test

import (
	"testing"

	"github.com/paroki/domus/internal/umat/domain"
	"github.com/paroki/domus/internal/umat/service"
	"github.com/stretchr/testify/require"
)

func TestMaskPII(t *testing.T) {
	require.Equal(t, "", domain.MaskPII(""))
	require.Equal(t, "****12", domain.MaskPII("12"))
	require.Equal(t, "****34", domain.MaskPII("1234"))
	require.Equal(t, "************3456", domain.MaskPII("3174010101903456"))
}

func TestValidatePII(t *testing.T) {
	got, err := service.ValidatePII("nik", "  3174010101903456  ")
	require.NoError(t, err)
	require.Equal(t, "3174010101903456", got)

	got, err = service.ValidatePII("nik", "   ")
	require.NoError(t, err)
	require.Equal(t, "", got)

	_, err = service.ValidatePII("nik", "123")
	require.Error(t, err)

	_, err = service.ValidatePII("nik", "123456789012345678901234567890123")
	require.Error(t, err)

	_, err = service.ValidatePII("nik", "1234\n5678")
	require.Error(t, err)
}

func TestNormalizePhoneAndEmail(t *testing.T) {
	p, err := service.NormalizePhone("0812-3456-7890")
	require.NoError(t, err)
	require.Equal(t, "+6281234567890", p)

	_, err = service.NormalizePhone("invalid")
	require.Error(t, err)

	e, err := service.NormalizeEmail("  Budi@Paroki.ORG ")
	require.NoError(t, err)
	require.Equal(t, "budi@paroki.org", e)

	_, err = service.NormalizeEmail("not-an-email")
	require.Error(t, err)
}

func TestDateValidation(t *testing.T) {
	_, _, err := service.ParseDateYYYYMMDD("tanggalLahir", "1990-05-12")
	require.NoError(t, err)

	_, _, err = service.ParseDateYYYYMMDD("tanggalLahir", "12-05-1990")
	require.Error(t, err)

	b := "2000-01-10"
	d := "1999-12-31"
	require.Error(t, service.ValidateDateOrder(&b, &d))
}
