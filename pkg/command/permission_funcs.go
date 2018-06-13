package command

var (
	// Anybody allows to any user
	Anybody = func(evaluate bool) (*PermissionStatus, error) {
		return &PermissionStatus{UserIsApproved: true, ApprovedRoles: []string{Anyone}}, nil
	}

	// AnyOf checks if any of the given permission checks is fulfilled
	AnyOf = func(permissionChecks ...PermissionCheck) PermissionCheck {
		return of(true, permissionChecks...)
	}

	// AllOf checks if all of the given permission checks are fulfilled
	AllOf = func(permissionChecks ...PermissionCheck) PermissionCheck {
		return of(false, permissionChecks...)
	}
)

// PermissionCheck represents any check of the user's permissions and returns PermissionStatus that contains the result
type PermissionCheck func(evaluate bool) (*PermissionStatus, error)

// Not reverses the given permission check
var Not = func(restriction PermissionCheck) PermissionCheck {
	return func(evaluate bool) (*PermissionStatus, error) {
		status, err := restriction(evaluate)
		reversedStatus := &PermissionStatus{User: status.User}
		reversedStatus.UserIsApproved = !status.UserIsApproved
		reversedStatus.RejectedRoles = status.ApprovedRoles
		reversedStatus.ApprovedRoles = status.RejectedRoles
		return reversedStatus, err
	}
}

func of(any bool, permissionChecks ...PermissionCheck) PermissionCheck {
	return func(evaluate bool) (*PermissionStatus, error) {
		statuses := make([]*PermissionStatus, 0, len(permissionChecks))
		for _, checkPermission := range permissionChecks {

			status, err := checkPermission(evaluate)
			if status.UserIsApproved && any {
				evaluate = false
			}
			statuses = append(statuses, status)
			if err != nil {
				return Flatten(statuses, any), err
			}
		}
		return Flatten(statuses, any), nil
	}
}

// Flatten takes a slice of permission statuses and returns one with the flattened values.
// anyOff parameter sets if the user should be approved in any of the given permission statuses or in all of them
func Flatten(statuses []*PermissionStatus, anyOff bool) *PermissionStatus {
	if len(statuses) == 0 {
		return &PermissionStatus{User: Unknown, UserIsApproved: true, ApprovedRoles: []string{Anyone}}
	}
	flattenedStatus := statuses[0]
	if len(statuses) == 1 {
		return flattenedStatus
	}
	for i := 1; i < len(statuses); i++ {
		status := statuses[i]
		flattenedStatus.User = status.User
		if anyOff {
			flattenedStatus.UserIsApproved = flattenedStatus.UserIsApproved || status.UserIsApproved
		} else {
			flattenedStatus.UserIsApproved = flattenedStatus.UserIsApproved && status.UserIsApproved
		}
		flattenedStatus.ApprovedRoles = append(flattenedStatus.ApprovedRoles, status.ApprovedRoles...)
		flattenedStatus.RejectedRoles = append(flattenedStatus.RejectedRoles, status.RejectedRoles...)
	}
	return flattenedStatus
}
