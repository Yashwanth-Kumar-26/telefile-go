package utils

import "fmt"

// FormatBytes formats bytes into a human-readable string
func FormatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// MakeProgressBar generates a visual progress bar
func MakeProgressBar(used, total int64) string {
	if total <= 0 {
		return "Infinite"
	}
	percent := float64(used) / float64(total)
	if percent > 1 {
		percent = 1
	}

	barLength := 10
	filledLength := int(percent * float64(barLength))
	bar := ""

	for i := 0; i < barLength; i++ {
		if i < filledLength {
			bar += "■"
		} else {
			bar += "□"
		}
	}

	return fmt.Sprintf("%s %.1f%%", bar, percent*100)
}
