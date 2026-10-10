package core

import (
	"errors"
	"log"

	"github.com/google/uuid"
)

func GenerateID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		nr := errors.Join(ErrGenerateID, err)
		log.Fatal(nr.Error())
	}

	return id
}
