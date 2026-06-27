package claude

func ParseSession(path string) ([]RawEvent, error) {
	result, err := claudeAdapterInstance.ParseSession(path)
	if err != nil {
		return nil, err
	}
	return result.Events, nil
}
