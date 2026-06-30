package tddcheck_test

import (
	"testing"

	"github.com/lwmacct/260622-go-pkg-tddcheck/pkg/tddcheck"
)

func TestRules(t *testing.T) {
	tddcheck.Project{Root: "internal"}.Assert(t)
}

func TestWriteTDDCheckIndex(t *testing.T) {
	tddcheck.Project{Root: "internal"}.WriteDoc(t, "")
}
