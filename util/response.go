package util

type ResponseEntity struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func BuildResponseEntity(success bool, message string, data interface{}) ResponseEntity {
	return ResponseEntity{
		Success: success,
		Message: message,
		Data:    data,
	}
}
