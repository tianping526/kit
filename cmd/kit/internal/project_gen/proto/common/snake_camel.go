package common

// SnakeString XxYy to xx_yy , XxYY to xx_y_y
func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	underline := true
	for i := 0; i < len(s); i++ {
		d := s[i]
		if d >= 'A' && d <= 'Z' {
			d = d + 32
			if !underline {
				data = append(data, '_')
			}
		}
		if d == '_' {
			underline = true
		} else {
			underline = false
		}
		data = append(data, d)
	}
	return string(data[:])
}

// CamelString xx_yy to XxYx  xx_y_y to XxYY
func CamelString(s string) string {
	data := make([]byte, 0, len(s))
	underline := false
	uppercase := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if !uppercase && d >= 'A' && d <= 'Z' {
			uppercase = true
		}
		if d >= 'a' && d <= 'z' && (underline || !uppercase) {
			d = d - 32
			underline = false
			uppercase = true
		}
		if uppercase && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			underline = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}
