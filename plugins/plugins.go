// Package plugins imports core plugins which causes them to register.
package plugins

import (
	_ "github.com/nslaughter/carbon/plugins/env"
	_ "github.com/nslaughter/carbon/plugins/git"
	_ "github.com/nslaughter/carbon/plugins/pathrename"
	_ "github.com/nslaughter/carbon/plugins/shell"
	_ "github.com/nslaughter/carbon/plugins/template"
	_ "github.com/nslaughter/carbon/plugins/textreplace"
)

// ============================================================================
