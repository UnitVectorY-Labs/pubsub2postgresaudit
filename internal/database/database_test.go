package database

import "testing"

func TestValidateIdentifier_Valid(t *testing.T) {
	cases := []string{"public", "my_table", "Table1", "a", "abc_123"}
	for _, tc := range cases {
		if err := ValidateIdentifier(tc); err != nil {
			t.Errorf("ValidateIdentifier(%q) returned error: %v", tc, err)
		}
	}
}

func TestValidateIdentifier_Invalid(t *testing.T) {
	cases := []string{"", "my-table", "my table", "schema.table", "tbl;DROP", `"quoted"`, "tbl\n"}
	for _, tc := range cases {
		if err := ValidateIdentifier(tc); err == nil {
			t.Errorf("ValidateIdentifier(%q) expected error, got nil", tc)
		}
	}
}
