package controller

import (
	"github.com/topicus-education-ops/argocd-namespace-operator/pkg/controller/application"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, application.Add)
}
