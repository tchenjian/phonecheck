package checker

import "errors"

type PhoneChecker interface {
	PhoneExists(phone string) bool
	Close() error
}

func NewPhoneChecker(typ string, dataFile string) (PhoneChecker, error) {
	switch typ {
	case "bitmap":
		return NewBitmapChecker(dataFile)
	case "bloom":
		return NewBloomChecker(dataFile)
	case "sqlite":
		return NewSqliteChecker(dataFile)
	case "tree":
		return NewTreeChecker(dataFile)
	case "mixtree":
		return NewMixTreeChecker(dataFile)
	default:
		return nil, errors.New("unsupported checker type: " + typ)
	}
}
