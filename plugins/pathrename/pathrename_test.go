package pathrename_test

import (
	"testing"

	"github.com/nslaughter/carbon/plugins/pathrename"
)

// Required fields are necessary and sufficient for validator
// Required fields at this time:
// 		 - dir
// 		 - substitutions
func TestRequiredFields(t *testing.T) {
	cases := []struct {
		name      string
		spec      map[interface{}]interface{}
		shouldErr bool
	}{
		{
			name: "required fields present",
			spec: map[interface{}]interface{}{
				"dir": "/demo/dir",
				"substitutions": []pathrename.Substitution{
					{Old: "old", New: "new"},
				}},
			shouldErr: false,
		},
		{
			name: "missing dir field",
			spec: map[interface{}]interface{}{
				"dir": "",
				"substitutions": []pathrename.Substitution{
					{Old: "old", New: "new"},
				}},
			shouldErr: true,
		},
		{
			name: "missing subs field",
			spec: map[interface{}]interface{}{
				"dir":           "/demo/dir",
				"substitutions": []pathrename.Substitution{}},
			shouldErr: true,
		},
	}

	for i, v := range cases {
		a := pathrename.New()
		a.Set(v.spec)
		if err := a.Validate(); err != nil {
			if !v.shouldErr {
				t.Fatalf("case %d: should not err: %v", i, err)
			}
		}
	}
}
