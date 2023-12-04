module agent

replace metricsPoll => ../../internal/metricsPoll

go 1.21.4

require metricsPoll v0.0.0-00010101000000-000000000000 // indirect
