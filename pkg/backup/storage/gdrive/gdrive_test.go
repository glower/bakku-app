package gdrive

import "testing"

func TestRemovePath(t *testing.T) {
	testcases := []struct {
		Name         string
		AbsolutePath string
		RelativePath string
		ExpectedPath string
	}{
		{
			Name:         "Test 1: [bar/baz]",
			AbsolutePath: "/foo/bar/baz/cat.jpg",
			RelativePath: "baz/cat.jpg",
			ExpectedPath: "bar/baz",
		},
		{
			Name:         "Test 2: [bar]",
			AbsolutePath: "/foo/bar/cat.jpg",
			RelativePath: "cat.jpg",
			ExpectedPath: "bar",
		},
		{
			Name:         "Test 2: [bar]",
			AbsolutePath: "/cat.jpg",
			RelativePath: "cat.jpg",
			ExpectedPath: "",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			result := remotePath(tc.AbsolutePath, tc.RelativePath)
			if tc.ExpectedPath != result {
				t.Fatalf("Remote path is invalid, expected %s, got %s\n", tc.ExpectedPath, result)
			}
		})
	}
}
