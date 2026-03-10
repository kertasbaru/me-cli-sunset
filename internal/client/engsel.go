package client

// GetProfile fetches the user profile.
func (c *Client) GetProfile(accessToken, idToken string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"access_token":  accessToken,
		"app_version":   "8.9.0",
		"is_enterprise": false,
		"lang":          "en",
	}

	res, err := c.SendAPIRequest("api/v8/profile", payload, idToken, "POST")
	if err != nil {
		return nil, err
	}

	data, ok := res["data"].(map[string]interface{})
	if !ok {
		return nil, nil
	}
	return data, nil
}

// GetBalance fetches the balance and credit information.
func (c *Client) GetBalance(idToken string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise": false,
		"lang":          "en",
	}

	res, err := c.SendAPIRequest("api/v8/packages/balance-and-credit", payload, idToken, "POST")
	if err != nil {
		return nil, err
	}

	data, _ := res["data"].(map[string]interface{})
	if data != nil {
		if balance, ok := data["balance"].(map[string]interface{}); ok {
			return balance, nil
		}
	}
	return nil, nil
}

// GetFamily fetches package family data by family code.
func (c *Client) GetFamily(idToken, familyCode string, isEnterprise *bool, migrationType *string) (map[string]interface{}, error) {
	isEnterpriseList := []bool{false, true}
	migrationTypeList := []string{"NONE", "PRE_TO_PRIOH", "PRIOH_TO_PRIO", "PRIO_TO_PRIOH"}

	if isEnterprise != nil {
		isEnterpriseList = []bool{*isEnterprise}
	}
	if migrationType != nil {
		migrationTypeList = []string{*migrationType}
	}

	for _, mt := range migrationTypeList {
		for _, ie := range isEnterpriseList {
			payload := map[string]interface{}{
				"is_show_tagging_tab":    true,
				"is_dedicated_event":     true,
				"is_transaction_routine": false,
				"migration_type":         mt,
				"package_family_code":    familyCode,
				"is_autobuy":             false,
				"is_enterprise":          ie,
				"is_pdlp":                true,
				"referral_code":          "",
				"is_migration":           false,
				"lang":                   "en",
			}

			res, err := c.SendAPIRequest("api/v8/xl-stores/options/list", payload, idToken, "POST")
			if err != nil {
				continue
			}

			status, _ := res["status"].(string)
			if status != "SUCCESS" {
				continue
			}

			data, ok := res["data"].(map[string]interface{})
			if !ok {
				continue
			}

			pf, ok := data["package_family"].(map[string]interface{})
			if !ok {
				continue
			}

			name, _ := pf["name"].(string)
			if name != "" {
				return data, nil
			}
		}
	}

	return nil, nil
}

// GetFamilies fetches all package families for a category.
func (c *Client) GetFamilies(idToken, packageCategoryCode string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"migration_type":        "",
		"is_enterprise":         false,
		"is_shareable":          false,
		"package_category_code": packageCategoryCode,
		"with_icon_url":         true,
		"is_migration":          false,
		"lang":                  "en",
	}

	res, err := c.SendAPIRequest("api/v8/xl-stores/families", payload, idToken, "POST")
	if err != nil {
		return nil, err
	}

	status, _ := res["status"].(string)
	if status != "SUCCESS" {
		return nil, nil
	}

	data, _ := res["data"].(map[string]interface{})
	return data, nil
}

// GetPackage fetches package details by option code.
func (c *Client) GetPackage(idToken, packageOptionCode, packageFamilyCode, packageVariantCode string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_transaction_routine": false,
		"migration_type":         "NONE",
		"package_family_code":    packageFamilyCode,
		"family_role_hub":        "",
		"is_autobuy":             false,
		"is_enterprise":          false,
		"is_shareable":           false,
		"is_migration":           false,
		"lang":                   "en",
		"package_option_code":    packageOptionCode,
		"is_upsell_pdp":          false,
		"package_variant_code":   packageVariantCode,
	}

	res, err := c.SendAPIRequest("api/v8/xl-stores/options/detail", payload, idToken, "POST")
	if err != nil {
		return nil, err
	}

	data, _ := res["data"].(map[string]interface{})
	return data, nil
}

// GetAddons fetches addon packages for a package option.
func (c *Client) GetAddons(idToken, packageOptionCode string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise":       false,
		"lang":                "en",
		"package_option_code": packageOptionCode,
	}

	res, err := c.SendAPIRequest("api/v8/xl-stores/options/addons-pinky-box", payload, idToken, "POST")
	if err != nil {
		return nil, err
	}

	data, _ := res["data"].(map[string]interface{})
	return data, nil
}

// GetNotifications fetches all notifications.
func (c *Client) GetNotifications(idToken string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise": false,
		"lang":          "en",
	}

	return c.SendAPIRequest("api/v8/notification-non-grouping", payload, idToken, "POST")
}

// GetNotificationDetail fetches detail for a specific notification.
func (c *Client) GetNotificationDetail(idToken, notificationID string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise":   false,
		"lang":            "en",
		"notification_id": notificationID,
	}

	return c.SendAPIRequest("api/v8/notification/detail", payload, idToken, "POST")
}

// GetTransactionHistory fetches the transaction history.
func (c *Client) GetTransactionHistory(idToken string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise": false,
		"lang":          "en",
	}

	res, err := c.SendAPIRequest("payments/api/v8/transaction-history", payload, idToken, "POST")
	if err != nil {
		return nil, err
	}

	data, _ := res["data"].(map[string]interface{})
	return data, nil
}

// GetTieringInfo fetches loyalty tiering information.
func (c *Client) GetTieringInfo(idToken string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise": false,
		"lang":          "en",
	}

	res, err := c.SendAPIRequest("gamification/api/v8/loyalties/tiering/info", payload, idToken, "POST")
	if err != nil {
		return nil, err
	}

	data, _ := res["data"].(map[string]interface{})
	return data, nil
}

// GetQuotaDetails fetches quota details for sentry monitoring.
func (c *Client) GetQuotaDetails(idToken string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"is_enterprise":    false,
		"lang":             "en",
		"family_member_id": "",
	}

	return c.SendAPIRequest("api/v8/packages/quota-details", payload, idToken, "POST")
}

// DashboardSegments fetches dashboard segment data.
func (c *Client) DashboardSegments(accessToken, idToken string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"access_token": accessToken,
	}

	return c.SendAPIRequest("dashboard/api/v8/segments", payload, idToken, "POST")
}

// GetPaymentMethods fetches available payment methods for a transaction.
func (c *Client) GetPaymentMethods(idToken, tokenConfirmation, paymentTarget string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"payment_type":       "PURCHASE",
		"is_enterprise":      false,
		"payment_target":     paymentTarget,
		"lang":               "en",
		"is_referral":        false,
		"token_confirmation": tokenConfirmation,
	}

	res, err := c.SendAPIRequest("payments/api/v8/payment-methods-option", payload, idToken, "POST")
	if err != nil {
		return nil, err
	}

	status, _ := res["status"].(string)
	if status != "SUCCESS" {
		return nil, nil
	}

	data, _ := res["data"].(map[string]interface{})
	return data, nil
}

// Unsubscribe cancels a package subscription.
func (c *Client) Unsubscribe(idToken, quotaCode, productDomain, productSubscriptionType string) (bool, error) {
	payload := map[string]interface{}{
		"product_subscription_type": productSubscriptionType,
		"quota_code":                quotaCode,
		"product_domain":            productDomain,
		"is_enterprise":             false,
		"unsubscribe_reason_code":   "",
		"lang":                      "en",
		"family_member_id":          "",
	}

	res, err := c.SendAPIRequest("api/v8/packages/unsubscribe", payload, idToken, "POST")
	if err != nil {
		return false, err
	}

	code, _ := res["code"].(string)
	return code == "000", nil
}
