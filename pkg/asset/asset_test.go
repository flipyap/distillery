package asset

import "testing"

func TestAsset(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		expectType  Type
		expectScore int
	}{
		{"test", "Test", Unknown, 0},
		{"test.tar.gz", "Test", Archive, 0},
		{"test.tar.gz.asc", "Test", Signature, 0},
	}

	for _, c := range cases {
		asset := New(c.name, c.displayName, "linux", "amd64", "1.0.0")

		if asset.GetName() != c.name {
			t.Errorf("expected name to be %s, got %s", c.name, asset.GetName())
		}
		if asset.GetDisplayName() != c.displayName {
			t.Errorf("expected display name to be %s, got %s", c.displayName, asset.GetDisplayName())
		}
		if asset.Type != c.expectType {
			t.Errorf("expected type to be %d, got %d", c.expectType, asset.Type)
		}
		if asset.score != c.expectScore {
			t.Errorf("expected score to be %d, got %d", c.expectScore, asset.score)
		}
	}
}

func TestAssetScoring(t *testing.T) {
	cases := []struct {
		name        string
		displayName string
		scoringOpts *ScoreOptions
		expectType  Type
		expectScore int
	}{
		{"test.tar.gz", "Test", &ScoreOptions{OS: []string{"linux"}, Arch: []string{"amd64"}, Extensions: []string{".tar.gz"}}, Archive, 15},
		{"test_amd64.tar.gz", "Test", &ScoreOptions{OS: []string{"linux"}, Arch: []string{"amd64"}, Extensions: []string{".tar.gz"}}, Archive, 20},
		{"test_linux_amd64.tar.gz", "Test", &ScoreOptions{OS: []string{"linux"}, Arch: []string{"amd64"}, Extensions: []string{"unknown", ".tar.gz"}}, Archive, 30},
		{"test_linux_x86_64.tar.gz", "Test", &ScoreOptions{OS: []string{"Linux"}, Arch: []string{"amd64", "x86_64"}, Extensions: []string{"unknown", ".tar.gz"}}, Archive, 30},
	}

	for _, c := range cases {
		asset := New(c.name, c.displayName, "linux", "amd64", "1.0.0")
		asset.Score(c.scoringOpts)

		if asset.GetName() != c.name {
			t.Errorf("expected name to be %s, got %s", c.name, asset.GetName())
		}
		if asset.GetDisplayName() != c.displayName {
			t.Errorf("expected display name to be %s, got %s", c.displayName, asset.GetDisplayName())
		}
		if asset.Type != c.expectType {
			t.Errorf("expected type to be %d, got %d", c.expectType, asset.Type)
		}
		if asset.score != c.expectScore {
			t.Errorf("expected score to be %d, got %d", c.expectScore, asset.score)
		}
	}
}
