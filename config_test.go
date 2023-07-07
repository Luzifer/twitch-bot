package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthTokenValidate(t *testing.T) {
	assert.NoError(
		t,
		configAuthToken{Hash: "$argon2id$v=19$m=47104,t=1,p=1$c2FsdA$vvmJ4o4I75omtCuCUcFupw"}.
			validate("21a84dfe-155f-427d-add2-2e63dab7502f"),
		"valid argon2id token",
	)
	assert.Error(
		t,
		configAuthToken{Hash: "$argon2id$v=19$m=47104,t=1,p=1$c2FsdA$vvmJ4o4I75omtCuCUcFupw"}.
			validate("a8188a1f-d7ff-4628-b0f3-0efa0ef364d8"),
		"invalid argon2id token",
	)

	assert.NoError(
		t,
		configAuthToken{Hash: "$2a$10$xa3tzQheujq3nj/vJdzKIOzaPvirI6FFangBNDFXI8BME4rBMhNoG"}.
			validate("21a84dfe-155f-427d-add2-2e63dab7502f"),
		"valid bcrypt token",
	)
	assert.Error(
		t,
		configAuthToken{Hash: "$2a$10$xa3tzQheujq3nj/vJdzKIOzaPvirI6FFangBNDFXI8BME4rBMhNoG"}.
			validate("a8188a1f-d7ff-4628-b0f3-0efa0ef364d8"),
		"invalid bcrypt token",
	)
}
