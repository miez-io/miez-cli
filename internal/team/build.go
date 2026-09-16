package team

import teambuild "github.com/manuel/miez-cli/internal/build"

// BuildPackage compiles a team authoring package into its installable
// generated team index without requiring an initialized workspace.
func (s *Service) BuildPackage(root string) (teambuild.Plan, error) {
	plan, err := teambuild.Build(root)
	if err != nil {
		return teambuild.Plan{}, err
	}
	if err := teambuild.Apply(root, plan); err != nil {
		return teambuild.Plan{}, err
	}
	return plan, nil
}
