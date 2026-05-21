package validator

import "fmt"

// idMessageProvider is the default MessageProvider using Indonesian language.
type idMessageProvider struct{}

// DefaultMessageProvider returns the default Indonesian MessageProvider.
func DefaultMessageProvider() MessageProvider {
	return &idMessageProvider{}
}

// Message implements MessageProvider.
func (p *idMessageProvider) Message(tag string, param string) string {
	switch tag {
	case "required":
		return "Field ini wajib diisi."
	case "email":
		return "Harus berupa alamat email yang valid."
	case "min":
		return fmt.Sprintf("Minimal %s karakter.", param)
	case "max":
		return fmt.Sprintf("Maksimal %s karakter.", param)
	case "len":
		return fmt.Sprintf("Harus tepat %s karakter.", param)
	case "oneof":
		return fmt.Sprintf("Nilai harus salah satu dari: %s.", param)
	case "url":
		return "Harus berupa URL yang valid."
	case "uuid", "uuid3", "uuid4", "uuid5":
		return "Harus berupa UUID yang valid."
	case "numeric":
		return "Harus berupa angka."
	case "alpha":
		return "Hanya boleh berisi huruf."
	case "alphanum":
		return "Hanya boleh berisi huruf dan angka."
	case "gt":
		return fmt.Sprintf("Harus lebih besar dari %s.", param)
	case "gte":
		return fmt.Sprintf("Harus lebih besar atau sama dengan %s.", param)
	case "lt":
		return fmt.Sprintf("Harus lebih kecil dari %s.", param)
	case "lte":
		return fmt.Sprintf("Harus lebih kecil atau sama dengan %s.", param)
	case "eqfield":
		return fmt.Sprintf("Harus sama dengan field %s.", param)
	case "nefield":
		return fmt.Sprintf("Tidak boleh sama dengan field %s.", param)
	default:
		return "Validasi gagal pada field ini."
	}
}
