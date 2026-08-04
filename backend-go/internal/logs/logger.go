package logs

func record(level Level, logType string, agentID int64, agentName, handle, message string) {
	if defaultStore != nil {
		defaultStore.add(level, logType, agentID, agentName, handle, message)
	}
}

func Info(logType string, message string) {
	record(LevelInfo, logType, 0, "", "", message)
}

func AgentInfo(agentID int64, agentName, handle, logType string, message string) {
	record(LevelInfo, logType, agentID, agentName, handle, message)
}

func AgentWarn(agentID int64, agentName, handle, logType string, message string) {
	record(LevelWarn, logType, agentID, agentName, handle, message)
}

func AgentError(agentID int64, agentName, handle, logType, message string) {
	record(LevelError, logType, agentID, agentName, handle, message)
}
