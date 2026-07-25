// @ctx: system priority derivation — maps user_severity × category to system_priority
// @hlv:sec [INPUT_VALIDATION] — determines ticket priority from user input fields

package ticketapi

func DerivePriority(severity TicketUserSeverity, category TicketCategory) TicketSystemPriority {
	switch severity {
	case TicketSeverityCritical:
		if category == TicketCategoryBug {
			return TicketPriorityP0
		}
		return TicketPriorityP1
	case TicketSeverityHigh:
		if category == TicketCategoryBug || category == TicketCategoryPerformance {
			return TicketPriorityP1
		}
		return TicketPriorityP2
	case TicketSeverityMedium:
		return TicketPriorityP2
	case TicketSeverityLow:
		if category == TicketCategoryDocumentation || category == TicketCategoryQuestion {
			return TicketPriorityP3
		}
		return TicketPriorityP3
	default:
		return TicketPriorityP3
	}
}

func IsValidCategory(c TicketCategory) bool {
	switch c {
	case TicketCategoryBug, TicketCategoryPerformance, TicketCategoryQuestion,
		TicketCategoryFeature, TicketCategoryDocumentation, TicketCategoryAccess,
		TicketCategoryNeedsAnalysis:
		return true
	}
	return false
}

func IsValidSeverity(s TicketUserSeverity) bool {
	switch s {
	case TicketSeverityCritical, TicketSeverityHigh, TicketSeverityMedium, TicketSeverityLow:
		return true
	}
	return false
}

func IsValidSource(s TicketSource) bool {
	switch s {
	case TicketSourceManual, TicketSourceTelemetry:
		return true
	}
	return false
}

func IsValidChannel(c TicketChannel) bool {
	switch c {
	case TicketChannelUI, TicketChannelCLI, TicketChannelTelemetry:
		return true
	}
	return false
}

func IsValidTransition(from, to TicketStatus) bool {
	allowed, ok := ValidStatusTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
