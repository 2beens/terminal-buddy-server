package internal

type Response struct {
	Ok            bool   `json:"ok"`
	Message       string `json:"message"`
	DataJsonBytes []byte `json:"data_json_bytes"`
}
