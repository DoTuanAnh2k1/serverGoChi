package service

import "github.com/DoTuanAnh2k1/serverGoChi/pkg/store"

// AuthorizeDecision is the answer to "Can user X execute command Y on NE Z?"
// plus the minimal trace the UI/CLI needs to show *why*.
type AuthorizeDecision struct {
	Allowed       bool   `json:"allowed"`
	Reason        string `json:"reason"`
	UserExists    bool   `json:"user_exists"`
	UserEnabled   bool   `json:"user_enabled"`
	NeReachable   bool   `json:"ne_reachable"`   // via some ne_access_group
	CommandOnNe   bool   `json:"command_on_ne"`  // the (command_id) is registered on (ne_id)
	CommandExecAllowed bool `json:"command_exec_allowed"` // via some cmd_exec_group
}

// Authorize evaluates the flat v2 rule:
//   1. user X exists + enabled + not locked
//   2. command Y is registered against NE Z (command.ne_id == neID)
//   3. X ∈ some ne_access_group that contains Z
//   4. X ∈ some cmd_exec_group that contains Y
//
// Any failure denies. On success, Reason is empty.
func Authorize(username string, neID, commandID int64) (AuthorizeDecision, error) {
	d := AuthorizeDecision{}
	s := store.GetSingleton()

	u, err := s.GetUserByUsername(username)
	if err != nil {
		return d, err
	}
	if u == nil {
		d.Reason = "user not found"
		return d, nil
	}
	d.UserExists = true
	if !u.IsEnabled || u.LockedAt != nil {
		d.Reason = "user not active"
		return d, nil
	}
	d.UserEnabled = true

	cmd, err := s.GetCommandByID(commandID)
	if err != nil {
		return d, err
	}
	if cmd == nil {
		d.Reason = "command not found"
		return d, nil
	}
	if cmd.NeID != neID {
		d.Reason = "command not registered on this NE"
		return d, nil
	}
	d.CommandOnNe = true

	userNeGroups, err := s.ListNeAccessGroupsOfUser(u.ID)
	if err != nil {
		return d, err
	}
	neGroups, err := s.ListNeAccessGroupsOfNE(neID)
	if err != nil {
		return d, err
	}
	if !anyIntersect(userNeGroups, neGroups) {
		d.Reason = "no ne_access_group grants access to this NE"
		return d, nil
	}
	d.NeReachable = true

	userCmdGroups, err := s.ListCmdExecGroupsOfUser(u.ID)
	if err != nil {
		return d, err
	}
	cmdGroups, err := s.ListCmdExecGroupsOfCommand(commandID)
	if err != nil {
		return d, err
	}
	if !anyIntersect(userCmdGroups, cmdGroups) {
		d.Reason = "no cmd_exec_group grants this command"
		return d, nil
	}
	d.CommandExecAllowed = true

	d.Allowed = true
	return d, nil
}

func anyIntersect(a, b []int64) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	set := make(map[int64]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := set[v]; ok {
			return true
		}
	}
	return false
}
