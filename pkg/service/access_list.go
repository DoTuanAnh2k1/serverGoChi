package service

import (
	"errors"
	"net"
	"strings"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

var ErrAccessListInvalid = errors.New("access-list: invalid list_type or match_type")

func CreateAccessListEntry(e *db_models.UserAccessList) error {
	if e.ListType != db_models.AccessListTypeBlacklist && e.ListType != db_models.AccessListTypeWhitelist {
		return ErrAccessListInvalid
	}
	switch e.MatchType {
	case db_models.AccessListMatchUsername,
		db_models.AccessListMatchIPCidr,
		db_models.AccessListMatchEmailDomain:
	default:
		return ErrAccessListInvalid
	}
	if e.Pattern == "" {
		return ErrAccessListInvalid
	}
	return store.GetSingleton().CreateAccessListEntry(e)
}

func ListAccessListEntries(listType string) ([]*db_models.UserAccessList, error) {
	return store.GetSingleton().ListAccessListEntries(listType)
}

func DeleteAccessListEntry(id int64) error {
	return store.GetSingleton().DeleteAccessListEntryByID(id)
}

// EvaluateAccessList applies the blacklist/whitelist rules described in
// models/db_models/auth.go:
//   - any blacklist entry matching the identity → DENY
//   - otherwise, if any whitelist entry exists for a match_type the identity
//     has a value for, the identity must match at least one → else DENY
//   - whitelist with zero entries of a given match_type = "allow all"
// Returns (allowed, reason). Reason is populated on DENY only.
func EvaluateAccessList(username, ip, email string) (bool, string) {
	entries, err := store.GetSingleton().ListAccessListEntries("")
	if err != nil || len(entries) == 0 {
		return true, ""
	}

	for _, e := range entries {
		if e.ListType != db_models.AccessListTypeBlacklist {
			continue
		}
		if identityMatches(e, username, ip, email) {
			return false, "blacklisted: " + e.Reason
		}
	}

	whitelistByType := map[string][]*db_models.UserAccessList{}
	for _, e := range entries {
		if e.ListType == db_models.AccessListTypeWhitelist {
			whitelistByType[e.MatchType] = append(whitelistByType[e.MatchType], e)
		}
	}
	if len(whitelistByType) == 0 {
		return true, ""
	}

	for matchType, list := range whitelistByType {
		if !hasValueForMatchType(matchType, username, ip, email) {
			continue
		}
		ok := false
		for _, e := range list {
			if identityMatches(e, username, ip, email) {
				ok = true
				break
			}
		}
		if !ok {
			return false, "not on whitelist (" + matchType + ")"
		}
	}
	return true, ""
}

func hasValueForMatchType(matchType, username, ip, email string) bool {
	switch matchType {
	case db_models.AccessListMatchUsername:
		return username != ""
	case db_models.AccessListMatchIPCidr:
		return ip != ""
	case db_models.AccessListMatchEmailDomain:
		return email != ""
	}
	return false
}

func identityMatches(e *db_models.UserAccessList, username, ip, email string) bool {
	switch e.MatchType {
	case db_models.AccessListMatchUsername:
		return username != "" && strings.EqualFold(e.Pattern, username)
	case db_models.AccessListMatchIPCidr:
		return ipMatchesCIDR(ip, e.Pattern)
	case db_models.AccessListMatchEmailDomain:
		return email != "" && strings.HasSuffix(strings.ToLower(email), "@"+strings.ToLower(strings.TrimPrefix(e.Pattern, "@")))
	}
	return false
}

func ipMatchesCIDR(ip, cidr string) bool {
	if ip == "" || cidr == "" {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	if !strings.Contains(cidr, "/") {
		return parsed.Equal(net.ParseIP(cidr))
	}
	_, net_, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return net_.Contains(parsed)
}
