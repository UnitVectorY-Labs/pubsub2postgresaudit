package consumer

import "testing"

func TestParseSubscription_Valid(t *testing.T) {
	proj, sub, err := ParseSubscription("projects/my-project/subscriptions/my-sub")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proj != "my-project" {
		t.Errorf("project = %q, want %q", proj, "my-project")
	}
	if sub != "my-sub" {
		t.Errorf("subscription = %q, want %q", sub, "my-sub")
	}
}

func TestParseSubscription_Invalid(t *testing.T) {
	cases := []string{
		"",
		"my-sub",
		"projects/my-project",
		"projects/my-project/subscriptions",
		"projects/my-project/topics/my-topic",
		"projects//subscriptions/",
		"a/b/c/d",
	}
	for _, tc := range cases {
		_, _, err := ParseSubscription(tc)
		if err == nil {
			t.Errorf("ParseSubscription(%q) expected error, got nil", tc)
		}
	}
}
