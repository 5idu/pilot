package xgo

func IsChanClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}
