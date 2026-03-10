package client

// GetGroupData fetches circle group status data.
func (c *Client) GetGroupData(idToken string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise": false,
		"lang":          "en",
	}

	return c.SendAPIRequest("family-hub/api/v8/groups/status", payload, idToken, "POST")
}

// GetGroupMembers fetches members of a circle group.
func (c *Client) GetGroupMembers(idToken, groupID string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"group_id":      groupID,
		"is_enterprise": false,
		"lang":          "en",
	}

	return c.SendAPIRequest("family-hub/api/v8/members/info", payload, idToken, "POST")
}

// ValidateCircleMember validates a phone number for circle membership.
func (c *Client) ValidateCircleMember(idToken, msisdn string) (map[string]interface{}, error) {
	encryptedMSISDN, err := c.crypto.EncryptCircleMSISDN(msisdn)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"msisdn":        encryptedMSISDN,
		"is_enterprise": false,
		"lang":          "en",
	}

	return c.SendAPIRequest("family-hub/api/v8/members/validate", payload, idToken, "POST")
}

// InviteCircleMember invites a member to a circle group.
func (c *Client) InviteCircleMember(accessToken, idToken, msisdn, name, groupID, memberIDParent string) (map[string]interface{}, error) {
	encryptedMSISDN, err := c.crypto.EncryptCircleMSISDN(msisdn)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"access_token":    accessToken,
		"group_id":        groupID,
		"is_enterprise":   false,
		"members": []map[string]interface{}{
			{
				"msisdn": encryptedMSISDN,
				"name":   name,
			},
		},
		"lang":             "en",
		"member_id_parent": memberIDParent,
	}

	return c.SendAPIRequest("family-hub/api/v8/members/invite", payload, idToken, "POST")
}

// RemoveCircleMember removes a member from a circle group.
func (c *Client) RemoveCircleMember(idToken, memberID, groupID, memberIDParent string, isLastMember bool) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"member_id":        memberID,
		"group_id":         groupID,
		"is_enterprise":    false,
		"is_last_member":   isLastMember,
		"lang":             "en",
		"member_id_parent": memberIDParent,
	}

	return c.SendAPIRequest("family-hub/api/v8/members/remove", payload, idToken, "POST")
}

// AcceptCircleInvitation accepts a circle group invitation.
func (c *Client) AcceptCircleInvitation(accessToken, idToken, groupID, memberID string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"access_token":  accessToken,
		"group_id":      groupID,
		"member_id":     memberID,
		"is_enterprise": false,
		"lang":          "en",
	}

	return c.SendAPIRequest("family-hub/api/v8/groups/accept-invitation", payload, idToken, "POST")
}

// CreateCircle creates a new family circle group.
func (c *Client) CreateCircle(accessToken, idToken, parentName, groupName, memberMSISDN, memberName string) (map[string]interface{}, error) {
	encryptedMSISDN, err := c.crypto.EncryptCircleMSISDN(memberMSISDN)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"access_token": accessToken,
		"parent_name":  parentName,
		"group_name":   groupName,
		"is_enterprise": false,
		"members": []map[string]interface{}{
			{
				"msisdn": encryptedMSISDN,
				"name":   memberName,
			},
		},
		"lang": "en",
	}

	return c.SendAPIRequest("family-hub/api/v8/groups/create", payload, idToken, "POST")
}

// SpendingTracker fetches the spending tracker data for a family hub.
func (c *Client) SpendingTracker(idToken, parentSubsID, familyID string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise":  false,
		"parent_subs_id": parentSubsID,
		"family_id":      familyID,
		"lang":           "en",
	}

	return c.SendAPIRequest("gamification/api/v8/family-hub/spending-tracker", payload, idToken, "POST")
}

// GetBonusData fetches bonus data for a family hub.
func (c *Client) GetBonusData(idToken, parentSubsID, familyID string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise":  false,
		"parent_subs_id": parentSubsID,
		"family_id":      familyID,
		"lang":           "en",
	}

	return c.SendAPIRequest("gamification/api/v8/family-hub/bonus/list", payload, idToken, "POST")
}
