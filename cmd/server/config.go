package main

import "time"

const (
	collectGoRuntimeMetricsTimeout = 10 * time.Second
	serverMaxRequestBodySize       = 1024 * 1024 * 1024 * 8 // 8GB
	serverReadTimeout              = 10 * time.Minute
)
