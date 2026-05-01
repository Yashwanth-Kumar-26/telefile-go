package handlers

import (
	"fmt"
	"time"

	"github.com/Yashwanth-Kumar-26/telefile/internal/utils"
)

func formatFileSize(bytes int64) string {
	return utils.FormatBytes(bytes)
}

func timeUntil(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		return "Expired"
	}

	if d < time.Hour {
		return fmt.Sprintf("%dm left", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh left", int(d.Hours()))
	}
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%dd left", int(d.Hours()/24))
	}
	return fmt.Sprintf("%d days left", int(d.Hours()/24))
}
