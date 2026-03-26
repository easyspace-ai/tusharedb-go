package market

func NormalizeAdjust(adjust string) string {
	switch adjust {
	case "qfq", "hfq":
		return adjust
	default:
		return "none"
	}
}
