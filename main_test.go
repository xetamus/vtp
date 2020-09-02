package main

import "testing"
import "github.com/mch1307/vaultlib"

type vc struct {}
func (v vc) IsAuthenticated() bool { return true }

type vcSubstitute struct {
	vc
}
func (v vcSubstitute) GetSecret(string) (vaultlib.Secret, error) {
	return vaultlib.Secret{map[string]string{"value": "value"}, nil}, nil
}

func TestPerformSubstitution(t *testing.T) {
  initializeLogger(false)
	vaultCli = vcSubstitute{}

	testString := "key: (( path/to/secret ))"

	expectedResult := "key: value"
	subs := parseTokens(testString)
	result, err := performSubstitutions(testString, subs)

	if err != nil {
		t.Errorf("Failed to perform substitution: %v", err)
	}

	if (result != expectedResult) {
		t.Errorf("Error parsing secret, Expected: %s, Got: %s", expectedResult, result)
	}
}

func TestStripParens(t *testing.T) {
	testString := "(( path/to/secret ))"

	expectedResult := "path/to/secret"
	result := stripParens(testString)

	if (result != expectedResult) {
		t.Errorf("Error parsing secret, Expected: %s, Got: %s", expectedResult, result)
	}
}

func TestParseSecret(t *testing.T) {
	testString := "path/to/secret:key"

	expectedPath := "path/to/secret"
	expectedKey := "key"
	
	path, key := parseSecret(testString)

	if (path != expectedPath) {
		t.Errorf("Error extracting secret path, Expected: %s, Got: %s", expectedPath, path)
	}

	if (key != expectedKey) {
		t.Errorf("Error extracting secret key, Expected: %s, Got: %s", expectedKey, key)
	}
}

func TestParseSecretWithoutKey(t *testing.T) {
	testString := "path/to/secret"

	expectedPath := testString
	expectedKey := "value"
	
	path, key := parseSecret(testString)

	if (path != expectedPath) {
		t.Errorf("Error extracting secret path, Expected: %s, Got: %s", expectedPath, path)
	}

	if (key != expectedKey) {
		t.Errorf("Error extracting secret key, Expected: %s, Got: %s", expectedKey, key)
	}
}