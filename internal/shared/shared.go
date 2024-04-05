// Package contains application wide variables ( How do you like this, Rob Pike? ¯\_(ツ)_/¯ )

package shared

import (
	"go.uber.org/zap"
)

// Common logger of application
var Logger *zap.Logger
