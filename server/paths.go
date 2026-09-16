package main

import "time"

const (
	outputDir       = "server/output"
	problemsDir     = outputDir + "/problems"
	authDir         = outputDir + "/auth"
	authFile        = authDir + "/leetcode_auth.json"
	allProblemsFile = outputDir + "/all_problems.json"
	logFile         = outputDir + "/local_leetcode_logs.txt"

	// os.ModeDir carries no permission bits, so MkdirAll(path, os.ModeDir)
	// creates a 0000 directory on Unix. Use explicit permissions instead.
	dirPerm  = 0o755
	filePerm = 0o644

	authFreshnessDays  = 5
	cacheFreshnessDays = 5

	httpTimeout  = 30 * time.Second
	pollInterval = time.Second
	pollTimeout  = 60 * time.Second
)
