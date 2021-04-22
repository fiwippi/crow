package api

import "testing"

func TestSchemeValidation(t *testing.T) {
	var badScheme HTTPScheme = "bad-scheme"

	if !validScheme(HTTP) {
		t.Errorf("Scheme valid but seen as invalid: %s\n", HTTP)
	}

	if !validScheme(HTTPS) {
		t.Errorf("Scheme valid but seen as invalid: %s\n", HTTPS)
	}

	if validScheme(badScheme) {
		t.Errorf("Scheme invalid but seen as valid: %s\n", badScheme)
	}

	mc.ResetLastAccessed()

	_, err := mc.createAndSendRequest("GET", APIDomain, "/op/", "thread/570368.json", badScheme, false)
	if err != ErrInvalidScheme {
		t.Error("Invalid scheme seen as valid: " + badScheme)
	}

	_, err = mc.createAndSendRequest("GET", APIDomain, "/op/", "thread/570368.json", HTTP, false)
	if err == ErrInvalidScheme {
		t.Error("Valid scheme seen as invalid: " + HTTP)
	}

	_, err = mc.createAndSendRequest("GET", APIDomain, "/op/", "thread/570368.json", HTTPS, false)
	if err == ErrInvalidScheme {
		t.Error("Valid scheme seen as invalid: " + HTTPS)
	}
}
