package gobump

import (
	"testing"

	"go/parser"
	"go/token"
)

func TestBump(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "testdata/test1", nil, parser.Mode(0))
	if err != nil {
		t.Fatal(err)
	}

	conf := Config{
		MinorDelta: 1,
	}

	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			vers, err := conf.ProcessNode(fset, f)
			if err != nil {
				t.Errorf("got error: %s", err)
			}
			if _, ok := vers["version"]; !ok {
				t.Errorf("should detect `version`")
			}
			if _, ok := vers["VERSION"]; !ok {
				t.Errorf("should detect `VERSION`")
			}
			if vers["version"] != "1.1.0" {
				t.Errorf("expected %v: got %v", "1.1.0", vers["version"])
			}
			if vers["VERSION"] != "2.1.0" {
				t.Errorf("expected %v: got %v", "2.1.0", vers["VERSION"])
			}
		}
	}
}

func TestBumpedVersion(t *testing.T) {
	testCases := []struct {
		name            string
		conf            Config
		currentVersion  string
		expectedVersion string
		expectedError   bool
	}{
		{
			name:            "major delta",
			conf:            Config{MajorDelta: 2},
			currentVersion:  "1.2.3",
			expectedVersion: "3.0.0",
		},
		{
			name:            "minor delta",
			conf:            Config{MinorDelta: 2},
			currentVersion:  "1.2.3",
			expectedVersion: "1.4.0",
		},
		{
			name:            "patch delta",
			conf:            Config{PatchDelta: 2},
			currentVersion:  "1.2.3",
			expectedVersion: "1.2.5",
		},
		{
			name:            "exact version",
			conf:            Config{Exact: "2.3.4"},
			currentVersion:  "1.2.3",
			expectedVersion: "2.3.4",
		},
		{
			name:           "invalid exact version",
			conf:           Config{Exact: "xxx"},
			currentVersion: "1.2.3",
			expectedError:  true,
		},
		{
			name:            "version bump down",
			conf:            Config{Exact: "0.1.2"},
			currentVersion:  "1.2.3",
			expectedVersion: "0.1.2",
		},
		{
			name:            "check patch version up",
			conf:            Config{Exact: "1.2.4", CheckVersionUp: true},
			currentVersion:  "1.2.3",
			expectedVersion: "1.2.4",
		},
		{
			name:            "check minor version up",
			conf:            Config{Exact: "1.3.0", CheckVersionUp: true},
			currentVersion:  "1.2.3",
			expectedVersion: "1.3.0",
		},
		{
			name:            "check major version up",
			conf:            Config{Exact: "2.0.0", CheckVersionUp: true},
			currentVersion:  "1.2.3",
			expectedVersion: "2.0.0",
		},
		{
			name:           "error on bump down",
			conf:           Config{Exact: "1.2.2", CheckVersionUp: true},
			currentVersion: "1.2.3",
			expectedError:  true,
		},
		{
			name:           "error on same version",
			conf:           Config{Exact: "1.2.3", CheckVersionUp: true},
			currentVersion: "1.2.3",
			expectedError:  true,
		},
		{
			name:            "set from empty",
			conf:            Config{Exact: "0.1.2", CheckVersionUp: true},
			currentVersion:  "",
			expectedVersion: "0.1.2",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.conf.bumpedVersion(tc.currentVersion)
			if got != tc.expectedVersion {
				t.Errorf("expected %v: got %v", tc.expectedVersion, got)
			}
			if (err != nil) != tc.expectedError {
				if tc.expectedError {
					t.Errorf("expected an error but got no error")
				} else {
					t.Errorf("expected no error but got an error: %v", err)
				}
			}
		})
	}
}
