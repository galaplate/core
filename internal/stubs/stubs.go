// Package stubs embeds the code-generation templates used by the make:*
// console commands, so they work wherever core is built from (a sibling
// checkout, the module cache or a vendor directory).
package stubs

import "embed"

//go:embed all:commands all:dto all:factories all:jobs all:migrations all:models all:package all:policies all:providers all:scheduler all:seeders
var FS embed.FS
