package helpers

import (
	"fmt"
	"time"
)

func HumanAge(t time.Time) string {
	d := time.Since(t)
	if d.Hours() >= 24 {
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	}
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.0fh", d.Hours())
	}
	if d.Minutes() >= 1 {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.0fs", d.Seconds())
}
