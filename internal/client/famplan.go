package client

// GetFamilyPlanData fetches the family plan member information.
func (c *Client) GetFamilyPlanData(idToken string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"group_id":      0,
		"is_enterprise": false,
		"lang":          "en",
	}

	return c.SendAPIRequest("sharings/api/v8/family-plan/member-info", payload, idToken, "POST")
}

// ValidateFamplanMSISDN validates a phone number for family plan membership.
func (c *Client) ValidateFamplanMSISDN(idToken, msisdn string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"with_bizon":         true,
		"with_family_plan":   true,
		"is_enterprise":      false,
		"with_optimus":       true,
		"lang":               "en",
		"msisdn":             msisdn,
		"with_regist_status": true,
		"with_enterprise":    true,
	}

	return c.SendAPIRequest("api/v8/auth/check-dukcapil", payload, idToken, "POST")
}

// ChangeFamplanMember changes a family plan member in a specific slot.
func (c *Client) ChangeFamplanMember(idToken, parentAlias, alias string, slotID int, familyMemberID, newMSISDN string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"parent_alias":     parentAlias,
		"is_enterprise":    false,
		"slot_id":          slotID,
		"alias":            alias,
		"lang":             "en",
		"msisdn":           newMSISDN,
		"family_member_id": familyMemberID,
	}

	return c.SendAPIRequest("sharings/api/v8/family-plan/change-member", payload, idToken, "POST")
}

// RemoveFamplanMember removes a member from the family plan.
func (c *Client) RemoveFamplanMember(idToken, familyMemberID string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise":    false,
		"family_member_id": familyMemberID,
		"lang":             "en",
	}

	return c.SendAPIRequest("sharings/api/v8/family-plan/remove-member", payload, idToken, "POST")
}

// SetFamplanQuotaLimit sets the quota allocation for a family plan member.
func (c *Client) SetFamplanQuotaLimit(idToken string, originalAllocation, newAllocation int, familyMemberID string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise": false,
		"member_allocations": []map[string]interface{}{
			{
				"new_text_allocation":       0,
				"original_text_allocation":  0,
				"original_voice_allocation": 0,
				"original_allocation":       originalAllocation,
				"new_voice_allocation":      0,
				"message":                   "",
				"new_allocation":            newAllocation,
				"family_member_id":          familyMemberID,
				"status":                    "",
			},
		},
		"lang": "en",
	}

	return c.SendAPIRequest("sharings/api/v8/family-plan/allocate-quota", payload, idToken, "POST")
}

// ValidatePUK validates a PUK code for a phone number.
func (c *Client) ValidatePUK(msisdn, puk string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise": false,
		"puk":           puk,
		"is_enc":        false,
		"msisdn":        msisdn,
		"lang":          "en",
	}

	return c.SendAPIRequest("api/v8/infos/validate-puk", payload, "", "POST")
}

// Dukcapil performs DUKCAPIL registration verification.
func (c *Client) Dukcapil(msisdn, kk, nik string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"msisdn": msisdn,
		"kk":     kk,
		"nik":    nik,
		"lang":   "en",
	}

	return c.SendAPIRequest("api/v8/auth/regist/dukcapil", payload, "", "POST")
}
